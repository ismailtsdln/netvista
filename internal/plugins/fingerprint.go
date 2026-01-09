package plugins

import (
	"strings"

	"github.com/ismailtsdln/netvista/pkg/models"
	"github.com/ismailtsdln/netvista/pkg/signatures"
)

type FingerprintPlugin struct {
	signatures []signatures.FingerprintSig
}

func NewFingerprintPlugin(sigs []signatures.FingerprintSig) *FingerprintPlugin {
	return &FingerprintPlugin{
		signatures: sigs,
	}
}

func (f *FingerprintPlugin) Name() string {
	return "Technology Fingerprinting"
}

func (f *FingerprintPlugin) Execute(target *models.Target) error {
	techs := make(map[string]bool)

	// Generic headers
	if server := target.Metadata.Headers["Server"]; server != "" {
		techs[server] = true
	}
	if poweredBy := target.Metadata.Headers["X-Powered-By"]; poweredBy != "" {
		techs[poweredBy] = true
	}

	for _, sig := range f.signatures {
		matched := false

		// Check headers
		if sig.Header != nil {
			for k, v := range sig.Header {
				if val, ok := target.Metadata.Headers[k]; ok && strings.Contains(strings.ToLower(val), strings.ToLower(v)) {
					matched = true
					break
				}
			}
		}

		// Check title
		if !matched && sig.Title != "" && strings.Contains(strings.ToLower(target.Metadata.Title), strings.ToLower(sig.Title)) {
			matched = true
		}

		if matched {
			techs[sig.Name] = true
		}
	}

	for t := range techs {
		target.Metadata.Technology = append(target.Metadata.Technology, t)
	}

	return nil
}
