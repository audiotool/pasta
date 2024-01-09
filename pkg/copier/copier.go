package copier

import (
	"context"
)

type Copier interface {
	// Returns weather the copier is for a specific url or not
	Matches(url string) bool
	// This function downloads and saves all files that should later be
	// pasted to the "to" directory into the "TempDir" directory.
	//
	// Copy must retrun a serializable object that is later saved inside "pasta.result.yaml".
	// Look at SourceInfo for an example.
	Copy(ctx context.Context, config CopyConfig) (any, error)
}

type CopyConfig struct {
	// URL of the dependency
	URL string
	// Directory from which copy takes place from
	From string
	// returns true if file with path relative to Src should be copied
	Keep func(path string) bool
	// Custom copier options from the pasta.yaml
	Options map[string]string
	// TempDir contains the path to write all files.
	TempDir string
	// ClearTarget is true if the target directory should be cleared before copying
	ClearTarget bool
}
