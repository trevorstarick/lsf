package lsf

import (
	"bytes"
	"errors"
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

func (m *manager) walk(pfd int, p string) {
	defer m.pendingJobs.Done()

	// fd, err := openat(pfd, filepath.Base(p), flags, 0o777)
	fd, err := syscall.Open(p, flags, 0o777)
	if err != nil {
		if errors.Is(err, os.ErrPermission) {
			return
		}

		fmt.Fprintf(os.Stderr, "%v\n", p)
		panic(err)
	}

	defer syscall.Close(fd)

	if fd < 0 {
		return
	}

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
				m.out <- path
			case syscall.DT_DIR:
				m.pendingJobs.Add(1)
				go func() {
					m.queue <- job{fd, path}
				}()
			default:
				continue
			}
		}
	}
}
