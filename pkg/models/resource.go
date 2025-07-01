package models

import (
	"io"
	"net/http"
	"net/url"
	"path/filepath"
)

type Resource struct {
	PluginID string `json:"id"`
	Scheme   string `json:"scheme"`
	Body     string `json:"body"`
}

type Metadata struct {
	Name      string `yaml:"name,omitempty"`       // 参数为尽力提供，可能不存在
	Size      int64  `yaml:"size,omitempty"`       // 参数为尽力提供，可能不存在
	MediaType string `json:"media_type,omitempty"` // 参数为尽力提供，可能不存在
}

type ResourceProvider interface {
	Metadata(scheme, body string) (*Metadata, error)
	Reader(scheme, body string) (io.ReadCloser, error)
}

func HttpMetadata(client *http.Client, urlStr string) (*Metadata, error) {
	u, err := url.Parse(urlStr)
	if err != nil {
		return nil, err
	}
	resp, err := client.Head(urlStr)

	if err != nil {
		if resp != nil {
			resp.Body.Close()
		}
		return nil, err
	}
	defer resp.Body.Close()
	return &Metadata{
		Name:      filepath.Base(u.Path),
		Size:      resp.ContentLength,
		MediaType: resp.Header.Get("Content-Type"),
	}, nil
}

func HttpReader(client *http.Client, urlStr string) (io.ReadCloser, error) {
	resp, err := client.Get(urlStr)
	if err != nil {
		if resp != nil {
			resp.Body.Close()
		}
		return nil, err
	}
	return resp.Body, nil
}
