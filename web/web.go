package web

import "embed"

//go:embed templates/* assets/*
var AssetsFS embed.FS
