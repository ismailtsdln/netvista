package plugins

import (
	"strings"

	"github.com/ismailtsdln/netvista/pkg/models"
	"github.com/ismailtsdln/netvista/pkg/signatures"
)

type WafPlugin struct {
	sigs []signatures.WafSig
}

func NewWafPlugin(sigs []signatures.WafSig) *WafPlugin {
	return &WafPlugin{sigs: sigs}
}

func (p *WafPlugin) Name() string {
	return "WAF Detection"
}

func (p *WafPlugin) Execute(target *models.Target) error {
	for _, sig := range p.sigs {
		match := false

		// Check headers
		for k, v := range sig.Header {
			if val, ok := target.Metadata.Headers[k]; ok {
				if v == "" || strings.Contains(strings.ToLower(val), strings.ToLower(v)) {
					match = true
					break
				}
			}
		}

		// Check body if no header match yet
		if !match && sig.Body != "" {
			if strings.Contains(target.Metadata.Body, sig.Body) {
				match = true
			}
		}

		if match {
			// We can add it to Technology or a new field if we want, but tech is fine for now
			// To distinguish, maybe prefix it
			target.Metadata.Technology = append(target.Metadata.Technology, "WAF:"+sig.Name)
		}
	}
	return nil
}
