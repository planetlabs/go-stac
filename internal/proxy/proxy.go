package proxy

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httputil"
	"net/url"
	"path"
	"strconv"
	"strings"
	"sync"

	"github.com/go-logr/logr"
	"github.com/planetlabs/go-stac/crawler"
)

type Handler struct {
	noCORS       bool
	base         string
	logger       *logr.Logger
	mutex        *sync.RWMutex
	allowedPaths map[string]bool
}

var _ http.Handler = (*Handler)(nil)

func (h Handler) respond(rw http.ResponseWriter, statusCode int) {
	if !h.noCORS {
		rw.Header().Set("Access-Control-Allow-Origin", "*")
	}
	rw.WriteHeader(statusCode)
}

func (h Handler) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	method := req.Method
	if method != http.MethodGet && method != http.MethodOptions {
		h.logger.V(1).Info("proxy method not allowed", "method", method)
		h.respond(rw, http.StatusMethodNotAllowed)
		return
	}

	prefix := path.Join("/", h.base)
	if !strings.HasSuffix(prefix, "/") {
		prefix += "/"
	}
	if !strings.HasPrefix(req.URL.Path, prefix) {
		h.logger.V(1).Info("invalid proxy base path", "path", req.URL.Path)
		h.respond(rw, http.StatusNotFound)
		return
	}

	parts := strings.SplitN(strings.TrimPrefix(req.URL.Path, prefix), "/", 2)
	if len(parts) != 2 {
		h.logger.V(1).Info("unexpected proxy path", "path", req.URL.Path)
		h.respond(rw, http.StatusNotFound)
		return
	}

	scheme := parts[0]
	switch scheme {
	case "http", "https":
		u, err := url.Parse(scheme + "://" + parts[1])
		if err != nil {
			h.respond(rw, http.StatusNotFound)
			return
		}
		h.proxyHTTP(u, rw, req)
		return
	case "file":
		// TODO
	}

	h.logger.V(1).Info("unsupported proxy scheme", "scheme", scheme)
	rw.WriteHeader(http.StatusNotFound)
}

func (h *Handler) proxyHTTP(target *url.URL, rw http.ResponseWriter, req *http.Request) {
	if !h.isAllowedPath(req.URL) {
		h.logger.V(1).Info("proxy path not in allowed list", "path", req.URL.Path)
		h.respond(rw, http.StatusNotFound)
		return
	}
	proxy := &httputil.ReverseProxy{
		Director:       getRequestModifier(target),
		ModifyResponse: h.getResponseModifier(target, req),
	}
	proxy.ServeHTTP(rw, req)
}

func (h *Handler) isAllowedPath(u *url.URL) bool {
	h.mutex.RLock()
	allowed := h.allowedPaths[u.Path]
	h.mutex.RUnlock()
	return allowed
}

func (h *Handler) getResponseModifier(target *url.URL, req *http.Request) func(*http.Response) error {
	return func(res *http.Response) error {
		res.Header.Del("Set-Cookie")
		if h.noCORS {
			res.Header.Del("Access-Control-Allow-Origin")
		} else {
			res.Header.Set("Access-Control-Allow-Origin", "*")
		}
		if res.StatusCode != http.StatusOK {
			return nil
		}

		if !strings.Contains(res.Header.Get("Content-Type"), "application/json") {
			return nil
		}

		data, readErr := readResponseBody(res)
		defer func() {
			updateResponseBody(res, data)
		}()

		if readErr != nil {
			return readErr
		}

		resource := crawler.Resource{}
		if err := json.Unmarshal(data, &resource); err != nil {
			return nil
		}

		h.mutex.Lock()
		defer h.mutex.Unlock()

		links := resource.Links()
		for _, link := range links {
			linkUrl, parseErr := url.Parse(link["href"])
			if parseErr != nil {
				return fmt.Errorf("failed to parse link %q: %w", link["href"], parseErr)
			}
			if !linkUrl.IsAbs() {
				linkUrl = target.ResolveReference(linkUrl)
			}

			newPath := h.GetPath(linkUrl)
			newUrl, err := url.Parse(newPath)
			if err != nil {
				return fmt.Errorf("failed to create url from %q: %w", newPath, err)
			}
			newUrl.Scheme = "http"
			newUrl.Host = req.Host
			newUrl.RawQuery = linkUrl.RawQuery
			link["href"] = newUrl.String()

			h.allowedPaths[newPath] = true
		}
		resource["links"] = links

		data, jsonErr := json.Marshal(resource)
		if jsonErr != nil {
			return fmt.Errorf("failed to encode resource as json: %w", jsonErr)
		}

		return nil
	}
}

func readResponseBody(res *http.Response) ([]byte, error) {
	if res.Header.Get("Content-Encoding") == "gzip" {
		reader, gzipErr := gzip.NewReader(res.Body)
		if gzipErr != nil {
			return nil, fmt.Errorf("failed to created gzip reader: %w", gzipErr)
		}
		data, readErr := io.ReadAll(reader)
		if readErr != nil {
			return nil, fmt.Errorf("failed to read gzip body: %w", readErr)
		}
		return data, readErr
	}

	data, readErr := io.ReadAll(res.Body)
	if readErr != nil {
		return nil, fmt.Errorf("failed to read response body: %w", readErr)
	}
	if err := res.Body.Close(); err != nil {
		return nil, fmt.Errorf("failed to close response body: %w", err)
	}
	return data, nil
}

func updateResponseBody(res *http.Response, data []byte) {
	res.Body = io.NopCloser(bytes.NewReader(data))
	res.ContentLength = int64(len(data))
	res.Header.Set("Content-Length", strconv.Itoa(len(data)))
	res.Header.Del("Content-Encoding")
}

type Options struct {
	NoCORS bool
	Base   string
	Entry  *url.URL
	Logger *logr.Logger
}

func New(options *Options) (*Handler, error) {
	if options.Entry == nil || !options.Entry.IsAbs() {
		return nil, fmt.Errorf("entry url must be absolute")
	}

	h := &Handler{
		base:         options.Base,
		logger:       options.Logger,
		mutex:        &sync.RWMutex{},
		allowedPaths: map[string]bool{},
	}
	h.allowedPaths[h.GetPath(options.Entry)] = true

	return h, nil
}

func (h *Handler) GetPath(u *url.URL) string {
	if u.Scheme == "file" {
		return path.Join("/", h.base, u.Scheme, u.Path)
	}
	return path.Join("/", h.base, u.Scheme, u.Host, u.Path)
}

func getRequestModifier(target *url.URL) func(*http.Request) {
	return func(req *http.Request) {
		req.URL = target
		req.Host = target.Host
		if _, ok := req.Header["User-Agent"]; !ok {
			req.Header.Set("User-Agent", "")
		}
		req.Header.Del("Cookie")
		req.Header.Del("Authorization")
	}
}
