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

	for _, path := range paths {
		go func() {
			for dirent := range c {
				fmt.Println(dirent)
			}
		}()

		err := lsf.WalkWithOptions(c, path, lsf.Options{
			MaxWorkers: runtime.NumCPU() * 8,
			Ignore: []string{
				".git",
				"node_modules",
			},
		})
		if err != nil {
			panic(err)
		}
	}

	close(c)
}
