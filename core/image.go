package core

import (
	"fmt"
	"io"
	"mime/multipart"
	"os"
)

type Image (ID)

func NewImage() Image {
	return Image(NewID())
}

func (image *Image) Save(raw multipart.File) error {
	dest, err := os.Create(fmt.Sprintf("media/images/%s", image))
	if err != nil {
		return Wrap(err, "error creating file destination")
	}
	_, err = io.Copy(dest, raw)
	if err != nil {
		return Wrap(err, "error copying to file destination")
	}
	return nil
}

func (image *Image) Delete() error {
	return os.Remove(fmt.Sprintf("media/images/%s", image))
}
