package tris

import (
	// "fmt"
	"io"
	"os"
)

func CopyFile(src string, dst string) error {
	s, err := os.Open(src)
	if err != nil {
		return err
	}
	defer s.Close()
	d, err := os.Create(dst)
	if err != nil {
		return err
	}
	if _, err := io.Copy(d, s); err != nil {
		d.Close()
		return err
	}
	return d.Close()
}

func PathExists(path string) (bool, error) {
	_, err := os.Stat(path)
	// fmt.Println("StatInfo:", statInfo)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}
