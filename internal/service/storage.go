package service

import (
	"io"
	"os"
	"path/filepath"
	"syscall"
	"time"
)

type FileStorage interface {
	UploadFileChunk(name string, data []byte) error
	ReadFile(name string) ([]byte, error)
	FilesData() ([]FileData, error)
}

type FileStorageService struct {
	path string
}

type FileData struct {
	Name       string
	UploadTime string
	UpdateTime string
}

func NewFileStorageService(path string) *FileStorageService {
	return &FileStorageService{
		path: path,
	}
}

func (fs *FileStorageService) UploadFileChunk(name string, data []byte) error {
	f, err := os.Create(filepath.Join(fs.path, filepath.Base(name)))
	if err != nil {
		return err
	}

	if _, err := f.Write(data); err != nil {
		return err
	}

	if err := f.Close(); err != nil {
		return err
	}

	return nil
}

func (fs *FileStorageService) ReadFile(name string) ([]byte, error) {
	f, err := os.Open(filepath.Join(fs.path, filepath.Base(name)))
	if err != nil {
		return nil, err
	}

	bytes, err := io.ReadAll(f)
	if err != nil {
		return nil, err
	}

	return bytes, nil
}

func (fs *FileStorageService) FilesData() ([]FileData, error) {
	files, err := os.ReadDir(fs.path)
	if err != nil {
		return nil, err
	}

	fds := make([]FileData, 0, len(files))

	for _, f := range files {
		info, err := os.Stat(filepath.Join(fs.path, f.Name()))
		if err != nil {
			return nil, err
		}

		fd := FileData{
			Name:       info.Name(),
			UpdateTime: info.ModTime().String(),
		}

		// Windows-specific :(
		sysInfo, ok := info.Sys().(*syscall.Win32FileAttributeData)
		if ok {
			fd.UploadTime = time.Unix(0, sysInfo.CreationTime.Nanoseconds()).String()
		}

		fds = append(fds, fd)
	}

	return fds, nil
}
