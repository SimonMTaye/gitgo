package main

import (
	"fmt"
	"io"
	"os"
	"path"
	"strconv"

	"github.com/SimonMTaye/gitgo/config"
	"github.com/SimonMTaye/gitgo/objects"
	"github.com/SimonMTaye/gitgo/repo"
	"github.com/teris-io/cli"
)

// Add command
var addHelp = "Usage: add [path]"
var add = cli.NewCommand("add", "stage a file").
	WithArg(cli.NewArg("path", "path of file to be staged")).
	WithAction(func(args []string, options map[string]string) int {
		if len(args) != 1 {
			fmt.Println(addHelp)
			return 1
		}
		cwd, err := os.Getwd()
		if err != nil {
			fmt.Println("error determining current working directory")
			fmt.Println(err.Error())
			return 1
		}
		repoDir, err := repo.FindRepo(cwd)
		if err != nil {
			fmt.Println(err.Error())
			return 1
		}
		err = AddHelper(repoDir, args[0])
		if err != nil {
			fmt.Println(err.Error())
			return 1
		}
		return 0
	})

// cat file command
var catFileHelp = "Usage cat-file [hash]"
var catFile = cli.NewCommand("cat-file", "display content of an object").
	WithArg(cli.NewArg("hash", "a uniquely identifiying hash")).
	WithOption(
		cli.NewOption("type", "show only the type of the object").
			WithChar('t').
			WithType(cli.TypeBool)).
	WithOption(
		cli.NewOption("size", "show only the size of the object").
			WithChar('s').
			WithType(cli.TypeBool)).
	WithAction(func(args []string, options map[string]string) int {
		if len(args) == 0 {
			fmt.Println(catFileHelp)
			return 1
		} else if len(args) > 1 {
			fmt.Println(catFileHelp)
			return 1
		}
		cwd, err := os.Getwd()
		if err != nil {
			fmt.Println(err)
		}
		repoDir, err := repo.FindRepo(cwd)
		if err != nil {
			fmt.Println(err)
		}
		obj, err := CatfileHelper(repoDir, args[0])
		if err != nil {
			fmt.Println(err)
		}
		if options["size"] == "true" {
			fmt.Println(obj.Size())
			// Print object size
			return 0
		} else if options["type"] == "true" {
			fmt.Println(obj.Type())
			return 0
		} else {
			fmt.Println(obj)
			return 0
		}
	})

var initCommand = cli.NewCommand("init", "create a new repository").
	WithAction(func(args []string, options map[string]string) int {
		cwd, err := os.Getwd()
		if err != nil {
			fmt.Println(err)
			return 1
		}
		err = repo.CreateRepo(cwd, "", "")
		if err != nil {
			fmt.Println(err)
			return 1
		}
		return 0
	})

var showRefCommand = cli.NewCommand("show-ref", "list all references in the repository").
	WithAction(func(args []string, options map[string]string) int {
		repoStruct, err := FindandOpenRepo()
		if err != nil {
			fmt.Println(err)
			return 1
		}
		refs, err := repoStruct.GetAllRefs()
		if err != nil {
			fmt.Println(err)
			return 1
		}
		for k, v := range refs {
			fmt.Printf("%s \t %s\n", k, v)
		}
		return 0
	})

var tagCommandHelp = "Usage: \n\ttag <tagname> <object-hash> [-m <message>]\n\ttag -d <tagname>"
var tagCommand = cli.NewCommand("tag", "tag an object with a name").
	WithOption(
		cli.NewOption("message", "message tag").
			WithChar('m').
			WithType(cli.TypeString)).
	WithOption(
		cli.NewOption("delete", "delete").
			WithChar('d').
			WithType(cli.TypeBool)).
	WithArg(
		cli.NewArg("tagname", "name of tag").
			WithType(cli.TypeString)).
	WithArg(
		cli.NewArg("hash", "the hash of the object or commit to tag").
			WithType(cli.TypeString).
			AsOptional()).
	WithAction(func(args []string, options map[string]string) int {
		repoStruct, err := FindandOpenRepo()
		if err != nil {
			fmt.Println(err)
			return 1
		}
		// Delete and message options can't be used together
		if options["message"] != "" && options["delete"] == "true" {
			fmt.Println(tagCommandHelp)
			return 1
		}
		// Delete an existing tag
		if options["delete"] == "true" {
			// Can't use object hash if trying to delete a tag
			if len(args) != 1 {
				fmt.Println(tagCommandHelp)
				return 1
			}
			// repo.DeleteTag
			err := repoStruct.DeleteTag(args[0])
			if err != nil {
				fmt.Println(err)
				return 1
			}
			return 0
		}
		// If not deleting, then we need two arguments, tagname + hash
		if len(args) != 2 {
			fmt.Println(tagCommandHelp)
			return 1
		}
		//Find object to be stored
		hash, err := repoStruct.FindObject(args[1])
		// Err means object being tagged doesn't exist
		if err != nil {
			fmt.Println(err)
			return 1
		}
		// Create a new tag
		if options["message"] != "" {
			// If there's a message, create a tag object then save a
			// tag reference that points to it
			obj, err := repoStruct.GetObject(hash)
			if err != nil {
				fmt.Println(err)
				return 1
			}
			tag := &objects.GitTag{}
			tag.SetObject(obj.Type(), hash)
			config, err := config.LoadConfig(path.Join(repoStruct.GitDir, "config"))
			if err != nil {
				fmt.Println(err)
				return 1
			}
			email, ok := (*config.All)["user"]["email"]
			if !ok {
				fmt.Println("No email set; please set the git user email")
				return 1
			}
			name, ok := (*config.All)["user"]["name"]
			if !ok {
				fmt.Println("No name set; please set the git user name")
				return 1
			}
			tag.SetTagger(name, email)
			err = repoStruct.SaveObject(tag)
			if err != nil {
				fmt.Println(err)
				return 1
			}
			err = repoStruct.SaveTag(args[0], objects.Hash(tag))
			if err != nil {
				fmt.Println(err)
				return 1
			}
			return 0
		} else {
			// If there's no message, just save a tag reference pointing
			// directly to the desired object
			err = repoStruct.SaveTag(args[0], hash)
			if err != nil {
				fmt.Println(err)
				return 1
			}
			return 0
		}
	})

//List files in the index (doesn't support other files for now)
var lsFilesCommand = cli.NewCommand("ls-files", "show information about files in the work tree or index").
	WithAction(func(args []string, options map[string]string) int {
		repoStruct, err := FindandOpenRepo()
		if err != nil {
			fmt.Println(err)
			return 1
		}
		index, err := repoStruct.Index()
		if err != nil {
			fmt.Println(err)
			return 1
		}
		for _, entry := range index.Entries {
			fmt.Println(entry.Name)
		}
		return 0
	})

// Compute object ID and optionally create a blob from a file
var hashObjectCommand = cli.NewCommand("hash-object", "Compute object ID and optionally create a blob from a file").
	WithOption(
		cli.NewOption("write", "write the object to the database").
			WithChar('w').
			WithType(cli.TypeBool)).
	WithOption(
		cli.NewOption("stdin", "read object from stdin").
			WithType(cli.TypeBool)).
	WithArg(
		cli.NewArg("path", "path to file to be hashed").
			AsOptional().
			WithType(cli.TypeString)).
	WithAction(func(args []string, options map[string]string) int {
		blob := &objects.GitBlob{}
		if options["stdin"] == "true" {
			data, err := io.ReadAll(os.Stdin)
			if err != nil {
				fmt.Println(err)
				return 1
			}
			blob.Deserialize(data)
		} else {
			if len(args[0]) != 1 {
				fmt.Println("fatal: must provide path to file or use --stdin")
				return 1
			}
			data, err := os.ReadFile(args[0])
			if err != nil {
				// If the string provided isn't an actual path
				// check if its a file name in the current directory
				cwd, err := os.Getwd()
				if err != nil {
					fmt.Println(err)
					return 1
				}
				file := path.Join(cwd, args[0])
				data, err = os.ReadFile(file)
				if err != nil {
					fmt.Println(err)
					return 1
				}
				blob.Deserialize(data)
			} else {
				// We successfully read the bytes from the path provided
				blob.Deserialize(data)
			}
		}
		// Print the hash
		fmt.Println(objects.Hash(blob))
		if options["write"] == "true" {
			repoStruct, err := FindandOpenRepo()
			if err != nil {
				fmt.Println(err)
				return 1
			}
			// Save the object if the "-w" is used
			repoStruct.SaveObject(blob)
		}
		return 0
	})

var commitCommand = cli.NewCommand("commit", "record changes to the repository").
	WithArg(cli.NewArg("message", "commit message")).
	WithAction(func(args []string, options map[string]string) int {
		repoStruct, err := FindandOpenRepo()
		if err != nil {
			fmt.Println(err)
			return 1
		}
		_, err = repoStruct.Index()
		if err != nil {
			// If there's a path error, then its likely the index doesn't exist yet and nothing has been added to the directory
			_, ok := err.(*os.PathError)
			if ok {
				fmt.Println("Nothing to commit")
				return 1
			}
			fmt.Println(err)
			return 1
		}

		err = repoStruct.Commit(args[0])
		if err != nil {
			fmt.Println(err)
			return 1
		}
		return 0
	})
var logCommand = cli.NewCommand("log", "Show commit logs").
	WithOption(
		cli.NewOption("branch", "branch to list").
			WithChar('b').
			WithType(cli.TypeString)).
	WithOption(
		cli.NewOption("commit-hash", "hash of commit").
			WithChar('c').
			WithType(cli.TypeString)).
	WithArg(
		cli.NewArg("distance", "number of commits to show").
			WithType(cli.TypeInt).
			AsOptional()).
	WithAction(func(args []string, options map[string]string) int {
		repoStruct, err := FindandOpenRepo()
		if err != nil {
			fmt.Println(err)
			return 1
		}
		specficCommit := options["commit-hash"]
		branch := options["branch"]
		// Will look for HEAD if branch and commit-hash aren't specified
		var startObj string
		if specficCommit != "" {
			startObj = specficCommit
		} else if branch != "" {
			startObj = branch
		} else {
			startObj = "HEAD"
		}

		distance := 5
		if len(args) == 1 {
			distance, err = strconv.Atoi(args[0])
			if err != nil {
				fmt.Println(err)
				return 1
			}
		}
		curHash := startObj
		for i := 0; i < distance; i++ {
			hash, err := repoStruct.FindObject(curHash)
			if err != nil {
				fmt.Println(err)
				return 1
			}
			obj, err := repoStruct.GetObject(hash)
			if err != nil {
				fmt.Println(err)
				return 1
			}
			commit, ok := obj.(*objects.GitCommit)
			if !ok {
				fmt.Println("Unexpected object found when following HEAD")
				return 1
			}
			fmt.Println(commit)
			curHash = commit.ParentHash
			if len(curHash) == 0 {
				break
			}
		}
		return 0
	})

// IMPROTANT: WILL ONLY REMOVE FILES FROM THE INDEX. THIS PROGRAM WON'T MODIFY THE WORKTREE
var rmCommand = cli.NewCommand("rm", "Remove files from the index").
	WithArg(
		cli.NewArg("path", "path to file to be removed").
			WithType(cli.TypeString)).
	WithAction(func(args []string, options map[string]string) int {
		repoStruct, err := FindandOpenRepo()
		if err != nil {
			fmt.Println(err)
			return 1
		}
		idx, err := repoStruct.Index()
		if err != nil {
			fmt.Println(err)
			return 1
		}
		exists, pos := idx.EntryExists(args[0])
		if exists {
			err = idx.DeleteEntry(pos)
			if err != nil {
				fmt.Println(err)
				return 1
			}

		} else {
			fmt.Println(args[0] + " doesn't exist. Use gitgo ls to list all files in the index")
			return 1
		}
		return 0
	})

var App = cli.New("small subset of git commands").
	WithCommand(add).
	WithCommand(catFile).
	WithCommand(rmCommand).
	WithCommand(logCommand).
	WithCommand(tagCommand).
	WithCommand(initCommand).
	WithCommand(commitCommand).
	WithCommand(showRefCommand).
	WithCommand(lsFilesCommand).
	WithCommand(hashObjectCommand)

func main() {
	os.Exit(App.Run(os.Args, os.Stdout))
}
