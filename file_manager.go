package main

import (
	"os"
	"syscall"
)

type FileManager struct {
	fp *os.File
}

func OpenFile(path string, mode int, lockToWrite bool) (*FileManager, error) {
	fd, err := syscall.Open(path, mode, 0644)

	if err != nil {
		return nil, err
	}

	var how int
	if lockToWrite {
		how = syscall.LOCK_EX
	} else {
		how = syscall.LOCK_SH
	}

	if err := syscall.Flock(fd, how); err != nil {
		syscall.Close(fd)

		return nil, err
	}

	return &FileManager{os.NewFile(uintptr(fd), path)}, nil
}

func (fm *FileManager) Read(b []byte) (int, error) {
	return fm.fp.Read(b)
}

func (fm *FileManager) Write(b []byte) (int, error) {
	return fm.fp.Write(b)
}

func (fm *FileManager) Seek(offset int64, whence int) (int64, error) {
	return fm.fp.Seek(offset, whence)
}

func (fm *FileManager) Close() error {
	syscall.Flock(int(fm.fp.Fd()), syscall.LOCK_UN)

	return fm.fp.Close()
}
