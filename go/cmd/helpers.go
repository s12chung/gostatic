package cmd

import (
	"archive/zip"
	"bufio"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

func promptStdIn(prompt string) (string, error) {
	reader := bufio.NewReader(os.Stdin)
	s := strings.TrimSuffix(strings.TrimSpace(prompt), ":") + ": "
	fmt.Print(s)
	return reader.ReadString('\n')
}

// from: https://stackoverflow.com/a/33853856/1090482
func downloadfile(url, dest string) error {
	out, err := os.Create(dest)
	if err != nil {
		return err
	}
	defer out.Close()

	// Get the data
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// Check server response
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("bad status: %s", resp.Status)
	}

	// Writer the body to file
	_, err = io.Copy(out, resp.Body)
	if err != nil {
		return err
	}

	return nil
}

// from: https://stackoverflow.com/a/24792688/1090482 and refactored
func unzip(src, dest string) error {
	r, err := zip.OpenReader(src)
	if err != nil {
		return err
	}
	defer func() {
		err = r.Close()
		if err != nil {
			panic(err)
		}
	}()

	err = os.MkdirAll(dest, 0750)
	if err != nil {
		return err
	}

	for _, f := range r.File {
		err := extractAndWriteFile(f, dest)
		if err != nil {
			return err
		}
	}

	return nil
}

func extractAndWriteFile(f *zip.File, dest string) error {
	rc, err := f.Open()
	if err != nil {
		return err
	}
	defer func() {
		err = rc.Close()
		if err != nil {
			panic(err)
		}
	}()

	path := filepath.Join(dest, f.Name)

	if f.FileInfo().IsDir() {
		err = os.MkdirAll(path, f.Mode())
		if err != nil {
			return err
		}
	} else {
		err = os.MkdirAll(filepath.Dir(path), f.Mode())
		if err != nil {
			return err
		}
		f, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
		if err != nil {
			return err
		}
		defer func() {
			err = f.Close()
			if err != nil {
				panic(err)
			}
		}()

		_, err = io.Copy(f, rc)
		if err != nil {
			return err
		}
	}
	return nil
}
