package main

import (
	"fmt"
	"os"

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

		lsf.Walk(c, path)
	}

	close(c)
}
