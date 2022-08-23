//go:build !darwin && !linux

package lsf

func walk(c chan string, pfd int, p string) {
	panic("this GOOS/GOARCH combo is not currently supported by lsf")
}
