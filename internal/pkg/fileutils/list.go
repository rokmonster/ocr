package fileutils

import "io/ioutil"

func GetFilesInDirectory(directory string) []string {
	files := []string{}

	dir, _ := ioutil.ReadDir(directory)

	for _, file := range dir {
		if !file.IsDir() {
			files = append(files, directory+"/"+file.Name())
		}
	}

	return files
}
