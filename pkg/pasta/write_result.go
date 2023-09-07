package pasta

import (
	"fmt"
	"os"
	"path"

	"gopkg.in/yaml.v3"
)

type yamlResult struct {
	URL        string `yaml:"url"`
	SourceInfo any    `yaml:"source_info,omitempty"`
	Skipped    bool   `yaml:"skipped,omitempty"`
	Error      string `yaml:"error,omitempty"`
}

type pastaResults struct {
	Deps []yamlResult `yaml:"deps"`
}

func writeResult(deps []Dependency, copyResults []CopyResult, parentDir string) error {
	// convert pasta.CopyResuts to yamlResults
	var results []yamlResult

	for i, result := range copyResults {
		url := deps[i].Option.URL

		var res yamlResult

		if result.Err != nil {
			fmt.Printf("error during copy of dependency %v: %v\n", i, result.Err)

			res = yamlResult{
				Error:   fmt.Sprintf("error during copy: %v", result.Err),
				Skipped: true,
				URL:     url,
			}
		} else {
			fmt.Printf("Copied files from %v\n", url)

			res = yamlResult{
				URL:        url,
				SourceInfo: result.CopierInfo,
			}
		}

		results = append(results, res)
	}

	// marshal CopyResult to yaml
	rescontent, err := yaml.Marshal(pastaResults{Deps: results})
	if err != nil {
		return fmt.Errorf("error at marshaling pasta.result.yaml: %v", err)
	}

	err = os.WriteFile(path.Join(parentDir, "pasta.result.yaml"), rescontent, 0644)

	if err != nil {
		return fmt.Errorf("error saving pasta.result.yaml: %v", err)
	}

	return nil
}
