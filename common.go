package lsf

import (
	"os"
	"path/filepath"
)

func Walk(c chan string, p string) {
	var err error
	p, err = filepath.Abs(p)
	if err != nil {
		panic(err)
	}

	fi, err := os.Open(filepath.Dir(p))
	if err != nil {
		panic(err)
	}

	pfd := fi.Fd()
	walk(c, int(pfd), p)
}
