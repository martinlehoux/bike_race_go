package core

import (
	"fmt"
	"io"
	"mime/multipart"
	"os"
)

type File (ID)

func NewFile() File {
	return File(NewID())
}

func (file File) Path() string {
	return fmt.Sprintf("media/files/%s", file)
}

func (file *File) Save(raw multipart.File) error {
	dest, err := os.Create(file.Path())
	if err != nil {
		return Wrap(err, "error creating file destination")
	}
	_, err = io.Copy(dest, raw)
	if err != nil {
		return Wrap(err, "error copying to file destination")
	}
	return nil
}

func (file *File) Delete() error {
	return os.Remove(file.Path())
}
