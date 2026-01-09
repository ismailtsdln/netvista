package plugins

import "github.com/ismailtsdln/netvista/pkg/models"

type Plugin interface {
	Name() string
	Execute(target *models.Target) error
}

type PluginManager struct {
	Plugins []Plugin
}

func NewPluginManager() *PluginManager {
	return &PluginManager{}
}

func (pm *PluginManager) Register(p Plugin) {
	pm.Plugins = append(pm.Plugins, p)
}

func (pm *PluginManager) RunAll(target *models.Target) {
	for _, p := range pm.Plugins {
		_ = p.Execute(target)
	}
}
