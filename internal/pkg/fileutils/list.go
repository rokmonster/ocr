package fileutils

import (
	"io/ioutil"
	"path/filepath"
)

func GetFilesInDirectory(directory string) []string {
	var files []string

	dir, _ := ioutil.ReadDir(directory)

	for _, file := range dir {
		if !file.IsDir() {
			files = append(files, filepath.Join(directory, file.Name()))
		}
	}

	return files
}
