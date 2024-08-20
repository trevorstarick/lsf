package lsf

import (
	"errors"
	"log/slog"
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

type Options struct {
	Logger *slog.Logger

	MaxWorkers int
	Ignore     []string
	// FollowSymlinks bool
	// MaxDepth       int
}

func WalkWithOptions(c chan string, p string, opts Options) error {
	if opts.MaxWorkers < 1 {
		opts.MaxWorkers = 1
	}

	var err error
	p, err = filepath.Abs(p)
	if err != nil {
		return err
	}

	fi, err := os.Open(filepath.Dir(p))
	if err != nil {
		return err
	}

	m := new(manager)
	m.pendingJobs = sync.WaitGroup{}
	m.out = c
	m.queue = make(chan job)

	errs := make(chan error)

	for i := 0; i < opts.MaxWorkers; i++ {
		go func() {
			var dupe bool

			for j := range m.queue {
				dupe = false

				baseDir := filepath.Base(j.p)

				for _, n := range opts.Ignore {
					match, err := filepath.Match(n, baseDir)
					if err != nil {
						errs <- err
					}

					if match {
						dupe = true

						if opts.Logger != nil {
							opts.Logger.Info("skipping directory",
								"dir", j.p,
								"rule", n,
							)
						}

						break
					}
				}

				if dupe {
					m.pendingJobs.Done()
					continue
				}

				err = m.walk(j.pd, j.p)
				if err != nil {
					errs <- err

					break
				}
			}
		}()
	}

	m.pendingJobs.Add(1)
	err = m.walk(int(fi.Fd()), p)
	if err != nil {
		return err
	}

	m.pendingJobs.Wait()

	close(errs)

	if len(errs) > 0 {
		e := []error{}
		for err := range errs {
			e = append(e, err)
		}

		return errors.Join(e...)
	}

	return nil
}

func Walk(c chan string, p string) error {
	return WalkWithOptions(c, p, Options{
		MaxWorkers: 1,
		Ignore:     []string{},
	})
}
