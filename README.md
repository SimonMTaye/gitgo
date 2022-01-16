# GitGo

gitgo is a *basic* git clone written in Go.
This project was written following the ["Write yourself a Git"](https://wyag.thb.lt) guide by Thibault Polge. Only git commands found in this guide will be implemented.

Most plumbing functionality is now compelete. This project will not focus on having a robust CLI interface; the focus is mostly on the plumbing commands
### Progress so far
#### Done
- [x] Find and open repos
- [x] Create and read git objects (trees, blobs, commits and t/ags
    
- [x] Parse index file (this file contains the data for the staging area)
- [x] Config manager for handling the different configuration locations and options
    - No advanced functionality, merely provides access to the configurations
- [x] Create tree/commit objects from the data in the index file
- [x] CLI commands
    - [x] add
    - [x] hash-object
    - [x] init
    - [x] cat-file
    - [x] ls-files
    - [x] show-ref
    - [x] tag
    - [x] commit
    - [x] log
    - [x] rm

#### Remaining
- [ ] Test that CLI commands work as expected
    - [x] ls-files
    - [x] init
    - [x] add
    - [x] hash-object
    - [x] cat-file
    - [x] show-ref
    - [x] log
    - [x] rm
    - [x] commit
    - [x] tag
  
### Known Bugs
- Git supports `0755` for file permission bits but `0644` is always used by gitgo
  - I have not been able to find documentation for when `0755` would be used
  - Linux builds often result in 'typechange' messages after using gitgo may be related to the permission bits when used to commit
- Bug when trying command on non-git dirs instead of git repo on Windows
  - The path "/" is weird on windows which is what is used to stop searching for repos on git
- Weird bug when modifying entries: entry with './' is added
  - Only appears when modifying, not first adding entry
  - Likely happens when an index entry is deleted
#### Will not be implemented
- Merging, managing remote repositories or otherwise interacting with other repos
    - While this part of git's core functionality, it is beyond the scope of this project
- Complex git configs (each repo has a config file; only the required information, such as branches, will be parsed. Other data that may impact how git works will be ignored.
- Index file extensions
- Support for symlinks and git-links
  - Checks based on os.IsRegular() will be used but this behavior will not be tested

#### Other Things to Keep in Mind
- When built for Windows, the index file might not function as expected when adding new entries. This is because the index file uses Inode, Device, Guid and Uuid numbers that do not exist (as far as I know) in Windows. Linux is unaffected


 ["Git from the Bottom Up"](https://jwiegley.github.io/git-from-the-bottom-up/) by John Wiegley was also immensely useful

