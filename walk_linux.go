package lsf

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"sync"
	"syscall"
	"unsafe"
)

const (
	flags              = syscall.O_RDONLY | syscall.O_DIRECTORY | syscall.O_CLOEXEC | syscall.O_NOATIME
	openatTrap uintptr = syscall.SYS_OPENAT

	// nameOffset is a compile time constant.
	nameOffset = int(unsafe.Offsetof(syscall.Dirent{}.Name))
)

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

func openat(dirfd int, path string, flags int, perm uint32) (int, error) {
	var p *byte
	p, err := syscall.BytePtrFromString(path)
	if err != nil {
		return 0, err
	}

	fd, _, errno := syscall.Syscall6(openatTrap, uintptr(dirfd), uintptr(unsafe.Pointer(p)), uintptr(flags), uintptr(perm), 0, 0)
	if errno != 0 {
		return 0, errno
	}

	return int(fd), nil
}

func nameFromDirent(de *syscall.Dirent) (name []byte) {
	// Because this GOOS' syscall.Dirent does not provide a field that specifies
	// the name length, this function must first calculate the max possible name
	// length, and then search for the NULL byte.
	ml := int(de.Reclen) - nameOffset

	// Convert syscall.Dirent.Name, which is array of int8, to []byte, by
	// overwriting Cap, Len, and Data slice header fields to the max possible
	// name length computed above, and finding the terminating NULL byte.
	sh := (*reflect.SliceHeader)(unsafe.Pointer(&name))
	sh.Cap, sh.Len = ml, ml
	sh.Data = uintptr(unsafe.Pointer(&de.Name[0]))

	if index := bytes.IndexByte(name, 0); index >= 0 {
		// Found NULL byte; set slice's cap and len accordingly.
		sh.Cap, sh.Len = index, index

		return
	}

	// NOTE: This branch is not expected, but included for defensive
	// programming, and provides a hard stop on the name based on the structure
	// field array size.
	sh.Cap, sh.Len = len(de.Name), len(de.Name)

	return
}

func walk(c chan string, pfd int, p string) {
	b := filepath.Base(p)
	fd, err := openat(pfd, b, flags, 0o777)
	if err != nil {
		panic(err)
	}

	defer syscall.Close(fd)

	if fd < 0 {
		return
	}

	// todo: make sync.Pool
	wg := sync.WaitGroup{}

	//nolint:forcetypeassert
	buf := bufPool.Get().([]byte)
	defer bufPool.Put(buf)

	//nolint:forcetypeassert
	dirent := direntPool.Get().(*syscall.Dirent)
	defer direntPool.Put(dirent)

	for {
		n, err := syscall.Getdents(fd, buf)
		if err != nil {
			panic(err)
		}

		if n <= 0 {
			break
		}

		for i := 0; i < n; {
			copy((*[unsafe.Sizeof(syscall.Dirent{})]byte)(unsafe.Pointer(dirent))[:], buf[i:n])
			i += int(dirent.Reclen)

			if dirent.Ino == 0 {
				continue
			}

			name := nameFromDirent(dirent)

			if len(name) == 0 || name[0] == '.' {
				continue
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
	}

	wg.Wait()
}
