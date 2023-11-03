package annotation

import (
	_ "image/gif"
	_ "image/png"
	_ "image/jpeg"
    "image"
    "os"
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
