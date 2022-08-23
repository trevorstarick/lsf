# lsf
faster recursive ls
---

`lsf` is a fast way to recursively list files in a directory and it's children. It acheives this by running concurrent directory traversals, and ignoring any ordering of the directory's contents


A significant portion of the development was aided by the following open-source projects: 
- https://cs.opensource.google/go/x/tools/+/master:internal/fastwalk/fastwalk.go 

- https://github.com/karrick/godirwalk 
