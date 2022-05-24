package normurl

import (
	"fmt"
	"net/url"
	"path/filepath"
	"runtime"
	"strings"
)

type Locator struct {
	url        *url.URL
	isFilepath bool
}

func (l *Locator) String() string {
	return l.url.String()
}

func (l *Locator) SetQueryParam(param string, value string) {
	if l.isFilepath {
		return
	}
	query := l.url.Query()
	if value != "" {
		query.Set(param, value)
	} else {
		query.Del(param)
	}
	l.url.RawQuery = query.Encode()
}

func (l *Locator) IsFilepath() bool {
	return l.isFilepath
}

func New(s string) (*Locator, error) {
	u, err := url.Parse(s)
	if err != nil {
		return nil, err
	}

	if u.Scheme == "" {
		if !filepath.IsAbs(s) {
			return nil, fmt.Errorf("expected absolute path")
		}
		loc := &Locator{
			url:        u,
			isFilepath: true,
		}
		return loc, nil
	}

	if u.Scheme == "file" {
		path := u.Path
		if runtime.GOOS == "windows" {
			path = filepath.FromSlash(strings.TrimPrefix(path, "/"))
		}
		u.Scheme = ""
		u.Path = path
		loc := &Locator{
			url:        u,
			isFilepath: true,
		}
		return loc, nil
	}

	if u.Scheme != "http" && u.Scheme != "https" {
		return nil, fmt.Errorf("unsupported scheme %s", u.Scheme)
	}

	return &Locator{url: u}, nil
}

func (base *Locator) Resolve(s string) (*Locator, error) {
	u, err := url.Parse(s)
	if err != nil {
		return nil, err
	}

	if u.Scheme != "" {
		return New(s)
	}

	if base.isFilepath {
		if filepath.IsAbs(s) {
			loc := &Locator{
				url:        u,
				isFilepath: true,
			}
			return loc, nil
		}

		baseDir := filepath.Dir(base.url.Path)
		path := filepath.Join(baseDir, s)
		loc := &Locator{
			url:        &url.URL{Path: path},
			isFilepath: true,
		}
		return loc, nil
	}

	resolved := base.url.ResolveReference(u)
	loc := &Locator{
		url:        resolved,
		isFilepath: false,
	}
	return loc, nil
}
