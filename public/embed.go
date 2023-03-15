package public

import (
	"embed"
	"io/fs"
	"net/http"
)

//go:embed build/*
var FS embed.FS

func StaticFS(relativePath string) http.FileSystem {
	sub, _ := fs.Sub(FS, "build"+relativePath)
	return http.FS(sub)
}
