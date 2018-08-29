package utils

import (
	"io"
	"io/ioutil"
	"os"
	"path"
	"sort"
	"strings"
)

func MkdirAll(path string) error {
	return os.MkdirAll(path, 0755)
}

func WriteFile(path string, bytes []byte) error {
	return ioutil.WriteFile(path, bytes, 0644)
}

// from: https://stackoverflow.com/a/21067803/1090482
func CopyFile(src, dst string) (err error) {
	in, err := os.Open(src)
	if err != nil {
		return
	}
	defer in.Close()
	out, err := os.Create(dst)
	if err != nil {
		return
	}
	defer func() {
		cerr := out.Close()
		if err == nil {
			err = cerr
		}
	}()
	if _, err = io.Copy(out, in); err != nil {
		return
	}
	err = out.Sync()
	return
}

func CleanFilePath(filePath string) string {
	filePath = strings.TrimLeft(filePath, ".")
	return strings.Trim(filePath, "/")
}

func ToSimpleQuery(queryMap map[string]string) string {
	queryArray := make([]string, len(queryMap))
	index := 0
	for key, value := range queryMap {
		queryArray[index] = key + "=" + value
		index += 1
	}
	sort.Strings(queryArray)
	return strings.Join(queryArray, "&")
}

func SliceList(slice []string) string {
	sliceLength := len(slice)
	newSlice := []string{}
	for index, item := range slice {
		newSlice = append(newSlice, item)

		if index == sliceLength-1 {
			continue
		}

		between := ", "
		if index == sliceLength-2 {
			between = " & "
		}
		newSlice = append(newSlice, between)
	}
	return strings.Join(newSlice, "")
}

func FilePaths(suffix string, dirPaths ...string) ([]string, error) {
	var filePaths []string

	for _, dirPath := range dirPaths {
		_, err := os.Stat(dirPath)
		if err != nil {
			return nil, err
		}
		files, err := ioutil.ReadDir(dirPath)
		if err != nil {
			return nil, err
		}

		for _, fileInfo := range files {
			if fileInfo.IsDir() || !strings.HasSuffix(fileInfo.Name(), suffix) {
				continue
			}
			filePaths = append(filePaths, path.Join(dirPath, fileInfo.Name()))
		}
	}
	return filePaths, nil
}
