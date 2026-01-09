package signatures

import (
	"os"

	"gopkg.in/yaml.v3"
)

type FingerprintSig struct {
	Name   string            `yaml:"name"`
	Header map[string]string `yaml:"header"`
	Title  string            `yaml:"title"`
	Body   string            `yaml:"body"`
}

type TakeoverSig struct {
	Name        string `yaml:"name"`
	Fingerprint string `yaml:"fingerprint"`
}

type Signatures struct {
	Fingerprints []FingerprintSig `yaml:"fingerprints"`
	Takeovers    []TakeoverSig    `yaml:"takeovers"`
}

func LoadSignatures(path string) (*Signatures, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var sigs Signatures
	err = yaml.Unmarshal(data, &sigs)
	if err != nil {
		return nil, err
	}

	return &sigs, nil
}
