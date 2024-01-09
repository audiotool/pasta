package cmd

import (
	"fmt"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/audiotool/pasta/pkg/copier"
	"github.com/audiotool/pasta/pkg/pasta"
	"gopkg.in/yaml.v3"
)

type pastaConf struct {
	KeepDirs bool          `yaml:"keep_dirs"`
	Deps     []*copierConf `yaml:"deps"`

	dependencies []pasta.Dependency
}

type copierConf struct {
	URL     string            `yaml:"url"`
	From    string            `yaml:"from"`
	To      string            `yaml:"to"`
	Include string            `yaml:"include"`
	Exclude string            `yaml:"exclude"`
	Options map[string]string `yaml:"options"`
	Files   []string          `yaml:"files"`
}

func (conf *copierConf) ToCopierOptions() (*copier.CopyConfig, error) {
	files := make(map[string]bool, len(conf.Files))
	for _, file := range conf.Files {
		files[file] = true
	}

	// prepare & compile include regexp
	includeRegexp, err := pasta.IncludeRegexp(conf.Include)

	if err != nil {
		return nil, err
	}

	// prepare & compile excludeRegexp regexp
	excludeRegexp, err := pasta.ExcludeRegexp(conf.Exclude)

	if err != nil {
		return nil, err
	}

	return &copier.CopyConfig{
		URL:  conf.URL,
		From: conf.From,
		Keep: func(path string) bool {
			if len(files) > 0 {
				return files[path]
			}
			return includeRegexp.MatchString(path) && !excludeRegexp.MatchString(path)
		},
		Options:     conf.Options,
		ClearTarget: len(files) == 0,
	}, nil
}

func newPastaConf(pathToYaml string) (*pastaConf, error) {
	// read yaml file
	yamlFile, err := os.ReadFile(pathToYaml)
	if err != nil {
		return nil, fmt.Errorf("couldn't read file %#v", err)
	}

	// unmarshal to pastaConf struct
	var c pastaConf
	err = yaml.Unmarshal(yamlFile, &c)
	if err != nil {
		return nil, fmt.Errorf("couldn't parse file, error: %v", err)
	}

	if err = c.validate(); err != nil {
		return nil, fmt.Errorf("invalid config: %v", err)
	}

	// pastaConf -> CopierOptions
	for i, config := range c.Deps {
		// convert pastaConf to CopierOptions
		option, _ := config.ToCopierOptions()

		// create tempdir
		option.TempDir, err = os.MkdirTemp("", "pasta")
		if err != nil {
			return nil, fmt.Errorf("couldn't create temp directory for dependency %v: %v", i, err)
		}

		dep := pasta.Dependency{
			Option: *option,
			Target: path.Join(path.Dir(pathToYaml), config.To),
		}

		c.dependencies = append(c.dependencies, dep)
	}

	return &c, nil
}

func (c *pastaConf) validate() error {
	var err error

	for i, config := range c.Deps {
		if config.URL == "" {
			return fmt.Errorf("dependency %v: 'url' is required", i)
		}

		err = validateFromField(config, i)
		if err != nil {
			return err
		}

		err = validateToField(config, i)
		if err != nil {
			return err
		}

		// files is exclusive with include/exclude
		if len(config.Files) > 0 && (config.Include != "" || config.Exclude != "") {
			return fmt.Errorf("dependency %v: files and include/exclude are mutually exclusive", i)
		}

		// convert pastaConf to CopierOptions
		_, err = config.ToCopierOptions()

		if err != nil {
			return fmt.Errorf("couldn't parse dependency %v: %v", i, err)
		}
	}

	return err
}

func validateFromField(config *copierConf, i int) error {
	// If From is `/`, we want to copy all files.
	// Since file paths dont start with `/`, we set empty string (so hasPrefix returns true).
	if config.From == "/" {
		config.From = ""
		return nil
	}

	if config.From == "" {
		return fmt.Errorf("dependency %v: 'from' is required", i)
	}

	if filepath.Clean(config.From)+"/" != config.From {
		return fmt.Errorf("dependency %v: 'from' must contain a clean path", i)
	}

	if !strings.HasSuffix(config.From, "/") {
		return fmt.Errorf("dependency %v: 'from' must end with '/'", i)
	}

	if strings.HasPrefix(config.From, "/") {
		return fmt.Errorf("dependency %v: 'from' must not start with '/'", i)
	}

	if config.From == "." {
		return fmt.Errorf("dependency %v: 'from' cannot be '.'", i)
	}
	return nil
}

func validateToField(config *copierConf, i int) error {
	if config.To == "." {
		if len(config.Files) == 0 {
			return fmt.Errorf("dependency %v: 'to' is set to '.' so 'files' must be used", i)
		}
		return nil
	}

	if config.To == "" {
		return fmt.Errorf("dependency %v: 'to' is required", i)
	}

	if !strings.HasSuffix(config.To, "/") {
		return fmt.Errorf("dependency %v: 'to' must end with '/'", i)
	}

	if filepath.Clean(config.To)+"/" != config.To {
		return fmt.Errorf("dependency %v: 'to' must contain a clean path", i)
	}

	if strings.HasPrefix(config.To, "/") {
		return fmt.Errorf("dependency %v: 'to' must not start with '/' or must be '.'", i)
	}

	return nil
}
