package lsf

import (
	"os"
	"path/filepath"
	"sync"
)

type job struct {
	pd int
	p  string
}

type manager struct {
	pendingJobs sync.WaitGroup
	out         chan string
	queue       chan job
}

func Walk(c chan string, p string, mx int) {
	if mx < 1 {
		mx = 1
	}

	var err error
	p, err = filepath.Abs(p)
	if err != nil {
		panic(err)
	}

	fi, err := os.Open(filepath.Dir(p))
	if err != nil {
		panic(err)
	}

	m := new(manager)
	m.pendingJobs = sync.WaitGroup{}
	m.out = c
	m.queue = make(chan job)

	for i := 0; i < mx; i++ {
		go func() {
			for j := range m.queue {
				m.walk(j.pd, j.p)
			}
		}()
	}

	m.pendingJobs.Add(1)
	m.walk(int(fi.Fd()), p)

	m.pendingJobs.Wait()
}
