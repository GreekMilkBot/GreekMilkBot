package models

import (
	"fmt"
	"io"
	"net/url"
	"strings"
)

type Resource url.URL

func (r *Resource) UnmarshalJSON(bytes []byte) error {
	u, err := url.Parse(strings.Trim(string(bytes), "\""))
	if err != nil {
		return err
	}
	*r = Resource(*u)
	return nil
}

func (r *Resource) MarshalJSON() ([]byte, error) {
	u := url.URL(*r)
	return []byte(fmt.Sprintf("\"%s\"", u.String())), nil
}

type Metadata struct {
	Name      string `yaml:"name,omitempty"`       // 参数为尽力提供，可能不存在
	Size      int64  `yaml:"size,omitempty"`       // 参数为尽力提供，可能不存在
	MediaType string `json:"media_type,omitempty"` // 参数为尽力提供
}

type ResourceProvider interface {
	Metadata(resource *Resource) (*Metadata, error)
	Reader(resource *Resource) (io.ReadCloser, error)
}
