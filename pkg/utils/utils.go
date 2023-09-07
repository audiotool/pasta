package utils

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path"
)

// CopyFile copies a file from src and tgt.
//
// If tgt ends with /, it is assumed to point to a directory, and the
// same file name from src is appended.
func CopyFile(src string, tgt string) error {
	if tgt[len(tgt)-1] == '/' {
		tgt = tgt + path.Base(src)
	}

	// open source file
	r, err := os.Open(src)
	if err != nil {
		return fmt.Errorf("error opening src: %v", err)
	}

	defer r.Close()

	// make parent dirs of target
	if err := os.MkdirAll(path.Dir(tgt), os.ModePerm); err != nil {
		return fmt.Errorf("error creating parents of tgt: %w", err)
	}

	// create target file
	w, err := os.Create(tgt)
	if err != nil {
		return fmt.Errorf("error creating tgt: %v", err)
	}
	defer w.Close()

	// copy
	_, err = io.Copy(w, r)

	if err != nil {
		return fmt.Errorf("error copying from src to tgt: %w", err)
	}

	return nil
}

func SaveFile(b []byte, tgt string) error {
	if err := os.MkdirAll(path.Dir(tgt), os.ModePerm); err != nil {
		return fmt.Errorf("error creating parents of tgt: %w", err)
	}

	f, err := os.Create(tgt)

	if err != nil {
		return fmt.Errorf("error creating file: %w", err)
	}

	defer f.Close()

	_, err = f.Write(b)

	if err != nil {
		return fmt.Errorf("could not write to file: %w", err)
	}

	return nil
}

// DownloadFile downloads a file from the given url and saves it to the given path.
func DownloadFile(url string, path string) error {
	// create target file
	out, err := os.Create(path)

	if err != nil {
		return fmt.Errorf("error creating target file: %w", err)
	}

	defer out.Close()

	// download the file via http and copy it to the target file
	fresp, err := http.Get(url)
	if err != nil {
		return fmt.Errorf("error downloading resource: %w", err)
	}

	_, err = io.Copy(out, fresp.Body)

	if err != nil {
		return fmt.Errorf("error copying file: %w", err)
	}

	fmt.Printf("Downloaded file %s to %s\n", url, path)

	return nil
}
