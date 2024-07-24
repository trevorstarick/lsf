//go:build !darwin && !linux

package lsf

import "errors"

func walk(c chan string, pfd int, p string) error {
	return errors.New("this GOOS/GOARCH combo is not currently supported by lsf")
}
