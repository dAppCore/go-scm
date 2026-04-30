// SPDX-License-Identifier: EUPL-1.2

package api

import (
	"embed"
	"io/fs"
	"net/http"
)

//go:embed ui/*
var uiAssets embed.FS

var embeddedUI = mustEmbeddedUI()
var embeddedIndexHTML = mustReadEmbeddedIndex()

func mustEmbeddedUI() http.FileSystem {
	sub, err := fs.Sub(uiAssets, "ui")
	if err != nil {
		panic(err)
	}
	return http.FS(sub)
}

func mustReadEmbeddedIndex() []byte {
	raw, err := uiAssets.ReadFile("ui/index.html")
	if err != nil {
		panic(err)
	}
	return raw
}
