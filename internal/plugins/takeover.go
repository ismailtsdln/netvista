package plugins

import (
	"strings"

	"github.com/ismailtsdln/netvista/pkg/models"
	"github.com/ismailtsdln/netvista/pkg/signatures"
)

type TakeoverPlugin struct {
	signatures []signatures.TakeoverSig
}

func NewTakeoverPlugin(sigs []signatures.TakeoverSig) *TakeoverPlugin {
	return &TakeoverPlugin{
		signatures: sigs,
	}
}

func (t *TakeoverPlugin) Name() string {
	return "Domain Takeover Detection"
}

func (t *TakeoverPlugin) Execute(target *models.Target) error {
	for _, sig := range t.signatures {
		if strings.Contains(target.Metadata.Body, sig.Fingerprint) {
			target.Metadata.Technology = append(target.Metadata.Technology, "Vulnerable: "+sig.Name)
		}
	}
	return nil
}
