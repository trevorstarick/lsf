package lsf_test

import (
	"log/slog"
	"os"
	"path/filepath"
	"testing"

	"github.com/trevorstarick/lsf"
)

func TestWalkWithOptions(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		c := make(chan string)
		p := "/tmp/tests/"

		os.MkdirAll(p, 0755)
		defer os.RemoveAll(p)

		for _, d := range []string{
			"node_modules",
			"node_modules/some_module",
			"applications/foo/bar/some.app",
			"projects/foo/bar/node_modules",
			"projects/foo/.git",
			"tmp",
			"tmpo",
			"test.app",
			"example",
		} {
			os.MkdirAll(filepath.Join(p, d), 0755)
		}

		for _, f := range []string{
			"example/file1.txt",
			"example/file2.txt",
			"tmp/debug.log",
			"tmp/trace.log",
			"tmpo/tmp.txt",
			"node_modules/some_module/index.js",
			"projects/foo/.git/index",
			"projects/foo/bar/node_modules/index.js",
			"test.app/some_file",
		} {
			os.Create(filepath.Join(p, f))
		}

		go func() {
			for dirent := range c {
				t.Logf("dirent: %s", dirent)
			}
		}()

		err := lsf.WalkWithOptions(c, p, lsf.Options{
			Logger:     slog.New(slog.NewTextHandler(os.Stdout, nil)),
			MaxWorkers: 1,
			Ignore: []string{
				"node_modules",
				"*.app",
				".git",
				"tmp",
			},
		})
		if err != nil {
			t.Errorf("WalkWithOptions() returned an error: %v", err)
		}
	})
}
