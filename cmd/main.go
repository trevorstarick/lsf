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

		err := lsf.WalkWithOptions(c, path, lsf.Options{
			MaxWorkers: runtime.NumCPU() * 8,
		})
		if err != nil {
			panic(err)
		}
	}

	close(c)
}
