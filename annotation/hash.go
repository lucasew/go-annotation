package annotation

import (
	"crypto/sha256"
	"fmt"
	"io"
	"os"
)

func HashFile(filepath string) (string, error) {
    f, err := os.Open(filepath)
    if err != nil {
        return "", err
    }
    hasher := sha256.New()
    _, err = io.Copy(hasher, f)
    if err != nil {
        return "", err
    }
    hash := fmt.Sprintf("%x", hasher.Sum(nil))
    return hash, nil
}
