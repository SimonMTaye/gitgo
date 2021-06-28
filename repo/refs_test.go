package repo

import (
    "os"
    "path"
    "testing"
    )


func CreateAndWrite(path string, content string) error {
    file, err := os.Create(path)
    if err != nil {
        return err
    }
    file.WriteString(content)
    file.Close()
    return nil
}


// Paths used when creating sample refs
const (
    firstRefPath = "refs/tags/first"
    secondRefPath = "refs/tags/second"
    thirdRefPath = "refs/heads/third"
    fourthRefPath = "refs/remotes/origin/fourth"
)
// Generate a directory tree that resembles a refs directory in a git project
func sampleRefsAndValues(rootDir string) error {
    refsDir := path.Join(rootDir, "refs")
    err := os.Mkdir(refsDir, DIR_FILEMODE)
    if err != nil {
        return err
    }
    // Create the directories usually found in refs
    err = os.Mkdir(path.Join(refsDir, "heads"), DIR_FILEMODE)
    if err != nil {
        return err
    }

    err = os.Mkdir(path.Join(refsDir, "tags"), DIR_FILEMODE)
    if err != nil {
        return err
    }

    err = os.MkdirAll(path.Join(refsDir, "remotes", "origin"), DIR_FILEMODE)
    if err != nil {
        return err
    }

    err = os.MkdirAll(path.Join(refsDir, "remotes", "backup"), DIR_FILEMODE)
    if err != nil {
        return err
    }

    err = CreateAndWrite(path.Join(rootDir, firstRefPath), "firstref")
    if err != nil {
        return err
    }

    err = CreateAndWrite(path.Join(rootDir, secondRefPath), "ref: " + firstRefPath)
    if err != nil {
        return err
    }

    err = CreateAndWrite(path.Join(rootDir,thirdRefPath), "ref: " + fourthRefPath)
    if err != nil {
        return err
    }

    err = CreateAndWrite(path.Join(rootDir, fourthRefPath), "fourthref")
    if err != nil {
        return err
    }
    return nil
}


func TestFindAllRefs(t *testing.T) {
    tmpDir := t.TempDir()
    gitDir := path.Join(tmpDir, ".git")
    err := os.Mkdir(gitDir, DIR_FILEMODE)
    if err != nil {
        t.Fatalf("Error creating test directory:\n%s", err.Error())
    }

    err = sampleRefsAndValues(gitDir)
    if err != nil {
        t.Fatalf("Error creating sample ref directories:\n%s", err.Error())
    }

    refMap, err := findAllRefs(gitDir)
    if err != nil{
        t.Fatalf("Unexpected error scanning for refs:\n%s", err.Error())
    }

    firstRef, ok := refMap[firstRefPath]
    if !ok {
        t.Errorf("Expected %s to exist", firstRefPath)
    } else if firstRef != "firstref" {
        t.Errorf("Expected %s to contain firstref, Got: %s", firstRefPath, firstRef)
    }

    secondRef, ok := refMap[secondRefPath]
    if !ok {
        t.Errorf("Expected %s to exist", secondRefPath)
    } else if secondRef != "firstref" {
        t.Errorf("Expected %s to contain firstref, Got: %s", secondRefPath,  secondRef)
    }

    thirdRef, ok := refMap[thirdRefPath]
    if !ok {
        t.Errorf("Expected %s to exist", thirdRefPath)
    } else if thirdRef != "fourthref" {
        t.Errorf("Expected %s to contain fourthref, Got: %s",thirdRefPath,  thirdRef)
    }

    fourthRef, ok := refMap[fourthRefPath]
    if !ok {
        t.Errorf("Expected %s to exist", fourthRefPath)
    } else if fourthRef != "fourthref" {
        t.Errorf("Expected %s to contain fourthref, Got: %s", fourthRefPath, fourthRef)
    }
}

func TestReadRef (t *testing.T) {
    tmpDir := t.TempDir()
    err := sampleRefsAndValues(tmpDir)
    if err != nil {
        t.Fatalf("Error creating sample directory and references:\n%s", err.Error())
    }
    firstRef, err := readRef(tmpDir, firstRefPath)
    if err != nil {
        t.Errorf("Unexpected error when reading %s:\n%s",firstRefPath, err.Error())
    } else if firstRef != "firstref" {
        t.Errorf("Expected %s to contain %s. Got: %s", firstRefPath, "firstref", firstRef)
    }

    secondRef, err := readRef(tmpDir, secondRefPath)
    if err != nil {
        t.Errorf("Unexpected error when reading %s:\n%s",secondRefPath, err.Error())
    } else if secondRef != "firstref" {
        t.Errorf("Expected %s to contain %s. Got: %s", secondRefPath, "firstref", secondRef)
    }

    thirdRef, err := readRef(tmpDir, thirdRefPath)
    if err != nil {
        t.Errorf("Unexpected error when reading %s:\n%s",thirdRefPath, err.Error())
    } else if thirdRef != "fourthref" {
        t.Errorf("Expected %s to contain %s. Got: %s", thirdRefPath, "fourthref", thirdRef)
    }

    fourthRef, err := readRef(tmpDir, fourthRefPath)
    if err != nil {
        t.Errorf("Unexpected error when reading %s:\n%s",fourthRefPath, err.Error())
    } else if thirdRef != "fourthref" {
        t.Errorf("Expected %s to contain %s. Got: %s", fourthRefPath, "fourthref", fourthRef)
    }
}

