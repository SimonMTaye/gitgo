# GitGo

gitgo is a *basic* git clone written in Go.
This project was written following the ["Write yourself a Git"](https://wyag.thb.lt) guide by Thibault Polge. Only git commands found in this guide will be implemented.

Most functionality is now compelete.
### Progress so far
#### Done
- [x] Find and open repos
- [x] Create and read git objects (trees, blobs, commits and tags
- [x] Parse index file (this file contains the data for the staging area)
#### Remaining
- [] Create tree/commit objects from the data in the index file
- [] Branches (fairly simple once committing has been implemented
- [] CLI commands
#### Will not be implemented
- Merging, managing remote repositories or otherwirse interacting with other repos
    - While this part of git's core functionality, it is beyond the scope of this project
- Complex git configs (each repo has a config file; only the required information, such as branches, will be parsed. Other data that may impact how git works will be ignored.
- Index file extensions


 ["Git from the Bottom Up"](https://jwiegley.github.io/git-from-the-bottom-up/) by John Wiegley was also immensely useful

