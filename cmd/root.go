package cmd

import (
	"context"
	_ "embed"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/audiotool/pasta/pkg/pasta"
	"github.com/spf13/cobra"
)

const pastayaml = "pasta.yaml"

var errCouldnFindPastaFile = errors.New("can't find '" + pastayaml + "'")

var (
	dryRunFlag  bool
	versionFlag bool
)

var (
	//go:embed version.txt
	version string
)

var RootCmd = &cobra.Command{
	Use:   "pasta",
	Short: "pasta - copy files between repositories",
	Run: func(cmd *cobra.Command, args []string) {
		if versionFlag {
			v := version
			if v == "" {
				v = "<unknown version>"
			}

			fmt.Println(v)

			os.Exit(0)
		}

		pathToYaml, err := findPastaFile()

		if err != nil {
			fmt.Fprintf(os.Stderr, "Error trying to find pasta file: %v\n", err)
			os.Exit(-1)
		}

		cfg, err := newPastaConf(pathToYaml)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error while parsing config: %v\n", err)
			os.Exit(-2)
		}

		if len(cfg.Deps) == 0 {
			fmt.Printf("No dependencies found in '%s'\n", pathToYaml)
			return
		}

		if dryRunFlag {
			fmt.Println("--dry-run is set, here's what would happen:")
			fmt.Println()
		}

		ctx := context.Background()
		err = pasta.Run(ctx, cfg.dependencies, dryRunFlag, cfg.KeepDirs, pathToYaml)

		if err != nil {
			fmt.Fprintf(os.Stderr, "Error while running pasta: %v\n", err)
			os.Exit(-4)
		}
	},
}

func Execute() {
	err := RootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

// findPastaFile tries to find pasta.yaml in the current or parent directories and returns the path
// to it. If it can't find it, it returns an error.
func findPastaFile() (string, error) {
	dir, err := os.Getwd()

	if err != nil {
		return "", fmt.Errorf("could not get working directory: %w", err)
	}

	dir, err = filepath.Abs(dir)

	if err != nil {
		return "", fmt.Errorf("could not get absolute path of working directory: %w", err)
	}

	root := filepath.VolumeName(dir)

	if root == "" {
		root = "/"
	}

	for {
		f := filepath.Join(dir, pastayaml)

		_, err = os.Stat(f)

		if err == nil {
			return f, nil
		}

		if !os.IsNotExist(err) {
			return "", fmt.Errorf("could not read file: %w", err)
		}

		if dir == root {
			return "", errCouldnFindPastaFile
		}

		dir = filepath.Dir(dir)
	}
}

func init() {
	RootCmd.Flags().BoolVar(&dryRunFlag, "dry-run", false, "don't do anything, just print what would be done")
	RootCmd.Flags().BoolVar(&versionFlag, "version", false, "shows the version of pasta")
}
