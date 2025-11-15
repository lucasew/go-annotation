package annotation

import (
	"crypto/sha256"
	"fmt"
	"github.com/google/uuid"
	"image"
	_ "image/gif"
	_ "image/jpeg"
	"image/png"
	"io"
	"os"
	"path"
)

func DecodeImage(filepath string) (image.Image, error) {
	f, err := os.Open(filepath)
	defer f.Close()
	if err != nil {
		return nil, err
	}
	m, _, err := image.Decode(f)
	if err != nil {
		return nil, err
	}
	return m, err
}

func IngestImage(img image.Image, outputDir string) error {
	tempFile := path.Join(outputDir, fmt.Sprintf("%s.png", uuid.New()))
	f, err := os.Create(tempFile)
	if err != nil {
		return err
	}
	hasher := sha256.New()
	w := io.MultiWriter(f, hasher)
	err = png.Encode(w, img)
	if err != nil {
		return err
	}
	err = f.Close()
	if err != nil {
		return err
	}
	err = os.Rename(tempFile, path.Join(outputDir, fmt.Sprintf("%x.png", hasher.Sum(nil))))
	if err != nil {
		os.Remove(tempFile)
		return err
	}
	return err
}
