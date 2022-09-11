package view

import (
	"embed"
	"io/fs"
	"net/http"
	"net/url"
	"path"
	"strings"

	"github.com/go-logr/logr"
)

//go:embed static
var embedded embed.FS

type Handler struct {
	handler http.Handler
	base    string
	logger  *logr.Logger
}

type Options struct {
	Base   string
	Logger *logr.Logger
}

var _ http.Handler = (*Handler)(nil)

func (h Handler) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	h.handler.ServeHTTP(rw, req)
}

func New(options *Options) (*Handler, error) {
	staticFS, err := fs.Sub(embedded, "static")
	if err != nil {
		return nil, err
	}

	prefix := path.Join("/", options.Base)
	if !strings.HasSuffix(prefix, "/") {
		prefix += "/"
	}

	staticHandler := http.FileServer(http.FS(staticFS))

	handler := &Handler{
		base:   options.Base,
		logger: options.Logger,
		handler: http.StripPrefix(prefix, http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
			if req.URL.Path == "" || strings.HasPrefix(req.URL.Path, "assets/") {
				staticHandler.ServeHTTP(rw, req)
				return
			}
			req.URL.Path = ""
			staticHandler.ServeHTTP(rw, req)
		})),
	}

	return handler, nil
}

func (h *Handler) GetPath(u *url.URL) string {
	if u.Scheme == "file" {
		return path.Join("/", h.base, u.Scheme, u.Path)
	}
	return path.Join("/", h.base, u.Scheme, u.Host, u.Path)
}
