package models

import (
	"fmt"
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
	Name      string `yaml:"name,omitempty"`
	Size      int64  `yaml:"size,omitempty"`
	MediaType string `json:"media_type,omitempty"`
}

type ResourceProvider interface {
	Metadata(resource *Resource)
}
