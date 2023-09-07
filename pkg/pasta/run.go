package pasta

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"strings"
	"sync"

	"github.com/audiotool/pasta/pkg/copier"
	"github.com/audiotool/pasta/pkg/github"
	"github.com/audiotool/pasta/pkg/utils"
	"golang.org/x/sync/errgroup"
	"gopkg.in/yaml.v3"
)

var copiers = []copier.Copier{&github.Copier{}}

type CopyResult struct {
	Err        error
	CopierInfo any
}

// tries to find matching copier, then executes copy with that copier
func executeCopy(ctx context.Context, option copier.CopyConfig) (any, error) {
	for _, c := range copiers {
		if !c.Matches(option.URL) {
			continue
		}

		res, err := c.Copy(ctx, option)
		if err != nil {
			return nil, fmt.Errorf("copy error: %v", err)
		}

		return res, nil
	}

	// return error if no copier was found
	return nil, fmt.Errorf("no copier found for url %v", option.URL)
}

type Dependency struct {
	Option copier.CopyConfig
	Target string
}

func copyToTemp(ctx context.Context, deps []Dependency) ([]CopyResult, error) {
	// dispatch goroutines copying files
	g, ctx := errgroup.WithContext(ctx)

	resultsMutex := sync.Mutex{}
	results := make([]CopyResult, len(deps))

	for i, dep := range deps {
		i := i
		dep := dep

		g.Go(func() error {
			res, err := executeCopy(ctx, dep.Option)

			resultsMutex.Lock()
			results[i] = CopyResult{Err: err, CopierInfo: res}
			resultsMutex.Unlock()

			return err
		})
	}

	err := g.Wait()

	if err != nil {
		return nil, err
	}

	return results, nil
}

func Run(ctx context.Context, deps []Dependency, dryRun, keepDirs bool, pastaFilePath string) (err error) {
	if !dryRun {
		defer func() {
			clearErr := clearTempDirs(deps)

			if clearErr != nil {
				err = errors.Join(err, clearErr)
			}
		}()
	}

	results, err := copyToTemp(ctx, deps)

	if err != nil {
		return fmt.Errorf("error copying dependencies: %v", err)
	}

	if !dryRun && !keepDirs {
		// remove all target directories
		for _, dep := range deps {
			if err := os.RemoveAll(dep.Target); err != nil {
				return fmt.Errorf("error removing target directory %v: %v", dep.Target, err)
			}
		}
	}

	// copy from temp to target
	if err := copyToTarget(deps, results, dryRun); err != nil {
		return fmt.Errorf("error copying files from temp to target dir: %v", err)
	}

	if !dryRun {
		if err := writeResult(deps, results, filepath.Dir(pastaFilePath)); err != nil {
			fmt.Printf("Error writing results file: %v", err)
			os.Exit(1)
		}
	}

	return clearTempDirs(deps)
}

func clearTempDirs(deps []Dependency) (err error) {
	for i, dep := range deps {
		if rmErr := os.RemoveAll(dep.Option.TempDir); err != nil {
			err = errors.Join(fmt.Errorf("error removing temp dir '%v' for dependency %v, error: %v", dep.Option.TempDir, i, rmErr))
		}
	}

	return err
}

// returns a list of files contained in root, as paths relative to root
func findFiles(root string) ([]string, error) {
	var res []string

	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if !d.IsDir() {
			rel, _ := filepath.Rel(root, path)
			res = append(res, rel)
		}

		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("error walking directory: %v", err)
	}

	return res, nil
}

func copyToTarget(deps []Dependency, results []CopyResult, dryRun bool) error {
	dep2Paths := make([][]string, len(deps))

	// retreive all paths, since they're used multiple times
	for i, dep := range deps {
		files, err := findFiles(dep.Option.TempDir)
		if err != nil {
			return fmt.Errorf("error listing files in temp directory %v: %v", dep.Option.TempDir, err)
		}
		dep2Paths[i] = files
	}

	// give dryRun output before checking for path uniqueness, makes debugging easier
	if dryRun {
		for i, dep := range deps {
			fmt.Printf("Dependency %v:\n", dep.Option.URL)

			if results[i].Err != nil {
				fmt.Printf(" Would output:\n")
				fmt.Println("   Error: ", strings.Join(strings.Split(results[i].Err.Error(), "\n"), "\n   "))
				continue
			}
			// else, print output & message
			fmt.Printf("  Would copy to %v:\n", dep.Target)
			for _, path := range dep2Paths[i] {
				fmt.Println("    *", path)
			}
			fmt.Printf("  Would output:\n")

			ci, err := yaml.Marshal(results[i].CopierInfo)

			if err != nil {
				return err
			}

			fmt.Println("   ", strings.Replace(string(ci), "\n", "\n    ", -1))

			fmt.Println()
		}

		if err := assertPathsUnique(dep2Paths); err != nil {
			fmt.Printf("Wouldn't copy: %v\n", err)
		}
		return nil
	}

	if err := assertPathsUnique(dep2Paths); err != nil {
		return fmt.Errorf("target paths are not unique: %v", err)
	}

	for i, dep := range deps {

		for _, p := range dep2Paths[i] {
			src := path.Join(dep.Option.TempDir, p)
			tgt := path.Join(dep.Target, p)

			// We copy the file (this also created the target directory) an os.Rename() would fail
			// if the source and target are on different devices/mounts. As we target only small
			// files, this should be fine.
			if err := utils.CopyFile(src, tgt); err != nil {
				fmt.Printf("Error copying file %v: %v", src, err)
			}
		}
	}

	return nil
}

// for a list [dependency-index][]paths, returns weather all paths are unique
func assertPathsUnique(dep2Paths [][]string) error {
	// make sure only unique paths exist
	pathSet := make(map[string]bool)
	for _, paths := range dep2Paths {
		for _, path := range paths {
			if pathSet[path] {
				return fmt.Errorf("two dependencies would copy to same file '%v'", path)
			}
			pathSet[path] = true
		}
	}
	return nil
}
