package plugins

import (
	"strings"

	"github.com/ismailtsdln/netvista/pkg/models"
)

type signature struct {
	name   string
	header map[string]string
	title  string
	body   string
}

type FingerprintPlugin struct {
	signatures []signature
}

func NewFingerprintPlugin() *FingerprintPlugin {
	return &FingerprintPlugin{
		signatures: []signature{
			{name: "WordPress", body: "/wp-content/", title: "WordPress"},
			{name: "Joomla", body: "/media/system/js/", title: "Joomla"},
			{name: "Drupal", body: "Drupal", header: map[string]string{"X-Generator": "Drupal"}},
			{name: "React", body: "node_modules", title: "React App"},
			{name: "Vue.js", body: "v-cloak"},
			{name: "Angular", body: "ng-version"},
			{name: "Nginx", header: map[string]string{"Server": "nginx"}},
			{name: "Apache", header: map[string]string{"Server": "Apache"}},
			{name: "Cloudflare", header: map[string]string{"Server": "cloudflare"}},
			{name: "PHP", header: map[string]string{"X-Powered-By": "PHP"}},
			{name: "ASP.NET", header: map[string]string{"X-Powered-By": "ASP.NET"}},
			{name: "Express", header: map[string]string{"X-Powered-By": "Express"}},
			{name: "Next.js", header: map[string]string{"X-Powered-By": "Next.js"}},
		},
	}
}

func (f *FingerprintPlugin) Name() string {
	return "Technology Fingerprinting"
}

func (f *FingerprintPlugin) Execute(target *models.Target) error {
	techs := make(map[string]bool)

	// Existing logic for generic headers
	if server := target.Metadata.Headers["Server"]; server != "" {
		techs[server] = true
	}
	if poweredBy := target.Metadata.Headers["X-Powered-By"]; poweredBy != "" {
		techs[poweredBy] = true
	}

	for _, sig := range f.signatures {
		matched := false

		// Check headers
		if sig.header != nil {
			for k, v := range sig.header {
				if val, ok := target.Metadata.Headers[k]; ok && strings.Contains(strings.ToLower(val), strings.ToLower(v)) {
					matched = true
					break
				}
			}
		}

		// Check title
		if !matched && sig.title != "" && strings.Contains(strings.ToLower(target.Metadata.Title), strings.ToLower(sig.title)) {
			matched = true
		}

		// Check body (not currently stored in target, but let's assume we might want to store a snippet or hash)
		// For now we'll skip body matching until we store it, or we rely on headers/title

		if matched {
			techs[sig.name] = true
		}
	}

	for t := range techs {
		target.Metadata.Technology = append(target.Metadata.Technology, t)
	}

	return nil
}
