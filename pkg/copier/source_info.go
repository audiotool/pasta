package copier

import "time"

// SourceInfo is serialized into the `pasta.result.yaml` in the end
type SourceInfo struct {
	Reference string `yaml:"reference,omitempty"`
	Message   string `yaml:"message,omitempty"`
	Author    Author `yaml:"author,omitempty"`
}

type Author struct {
	Date  time.Time `yaml:"date,omitempty"`
	Name  string    `yaml:"name,omitempty"`
	Email string    `yaml:"email,omitempty"`
}
