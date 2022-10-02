package main

import (
	"fmt"
	"os"
	"runtime"

	"github.com/trevorstarick/lsf"
)

func main() {
	c := make(chan string)

	paths := os.Args[1:]

	if len(paths) == 0 {
		paths = []string{"."}
	}

	lsf.AddToIgnoreList(".git")
	lsf.AddToIgnoreList("node_modules")

	for _, path := range paths {
		go func() {
			for dirent := range c {
				fmt.Println(dirent)
			}
		}()

		lsf.Walk(c, path, runtime.NumCPU()*8)
	}

	close(c)
}
