package screenshot

import (
	"bytes"
	"fmt"
	"image"
	_ "image/jpeg"
	_ "image/png"

	"github.com/corona10/goimagehash"
)

func GeneratePHash(imageBytes []byte) (string, error) {
	img, _, err := image.Decode(bytes.NewReader(imageBytes))
	if err != nil {
		return "", fmt.Errorf("failed to decode image: %v", err)
	}

	hash, err := goimagehash.PerceptionHash(img)
	if err != nil {
		return "", fmt.Errorf("failed to generate pHash: %v", err)
	}

	return hash.ToString(), nil
}

func HammingDistance(h1, h2 string) (int, error) {
	hash1, err := goimagehash.ImageHashFromString(h1)
	if err != nil {
		return -1, err
	}

	hash2, err := goimagehash.ImageHashFromString(h2)
	if err != nil {
		return -1, err
	}

	return hash1.Distance(hash2)
}
