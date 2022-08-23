package lsf

import (
	"os"
	"path/filepath"
	"sync"
	"syscall"
	"unsafe"
)

const flags = syscall.O_RDONLY | syscall.O_DIRECTORY | syscall.O_CLOEXEC

var (
	bufPool = sync.Pool{
		New: func() any {
			return make([]byte, os.Getpagesize())
		},
	}

	direntPool = sync.Pool{
		New: func() any {
			return new(syscall.Dirent)
		},
	}
)

//nolint:revive
//go:linkname readdir_r syscall.readdir_r
func readdir_r(dir uintptr, entry *syscall.Dirent, result **syscall.Dirent) (res syscall.Errno)

//go:linkname fdopendir syscall.fdopendir
func fdopendir(fd int) (dir uintptr, err error)

func walk(c chan string, pfd int, p string) {
	fd, err := syscall.Open(p, flags, 0o777)
	if err != nil {
		panic(err)
	}

	defer syscall.Close(fd)

	if fd < 0 {
		return
	}

	fdptr, err := fdopendir(fd)
	if err != nil {
		panic(err)
	}

	// todo: make sync.Pool
	wg := sync.WaitGroup{}

	//nolint:forcetypeassert since these are pools they're never not going to be []byte
	buf := bufPool.Get().([]byte)
	defer bufPool.Put(buf)

	//nolint:forcetypeassert since these are pools they're never not going to be *syscall.Dirent
	dirent := direntPool.Get().(*syscall.Dirent)
	defer direntPool.Put(dirent)

	//nolint:forcetypeassert since these are pools they're never not going to be *syscall.Dirent
	entptr := direntPool.Get().(*syscall.Dirent)
	defer direntPool.Put(entptr)

	for {
		if errno := readdir_r(fdptr, dirent, &entptr); errno != 0 {
			if errno == syscall.EINTR {
				continue
			}
			panic(errno)
		}

		if entptr == nil { // EOF
			break
		}

		if dirent.Ino == 0 {
			continue
		}

		name := (*[len(syscall.Dirent{}.Name)]byte)(unsafe.Pointer(&dirent.Name))[:]

		if dirent.Namlen == 0 || name[0] == '.' {
			continue
		}

		//
		for i, c := range name {
			if c == 0 {
				name = name[:i]

				break
			}
		}

		path := filepath.Join(p, string(name))

		switch dirent.Type {
		case syscall.DT_REG:
			c <- path
		case syscall.DT_DIR:
			wg.Add(1)
			go func() {
				walk(c, fd, path)
				wg.Done()
			}()
		default:
			continue
		}
	}

	wg.Wait()
}