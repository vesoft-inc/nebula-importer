package base

import (
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"strings"
)

func MustCreateFile(filePath string) *os.File {
	if err := os.MkdirAll(path.Dir(filePath), 0775); err != nil && !os.IsExist(err) {
		panic(err)
	}
	file, err := os.Create(filePath)
	if err != nil {
		panic(err)
	}
	return file
}

func PathFileList(path string) ([]string, error) {
	info, err := os.Stat(path)
	if err != nil {
		return nil, err
	}
	if !info.IsDir() {
		return []string{path}, nil
	}

	files, err := ioutil.ReadDir(path)
	if err != nil {
		return nil, err

	}

	var filenames []string
	for _, f := range files {
		if !f.IsDir() {
			filenames = append(filenames, filepath.Join(path, f.Name()))
		}
	}
	return filenames, nil
}

func PathExists(dir string) bool {
	_, err := os.Stat(dir)
	return !os.IsNotExist(err)
}

func DirExists(dir string) bool {
	info, err := os.Stat(dir)
	if os.IsNotExist(err) {
		return false
	}
	return info.IsDir()
}

func FileExists(filename string) bool {
	info, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
}

func IsValidType(t string) bool {
	switch strings.ToLower(t) {
	case "string":
	case "int":
	case "float":
	case "double":
	case "bool":
	case "timestamp":
	default:
		if strings.HasPrefix(t, "date-timestamp") && len(strings.Split(t, ":")) == 2 {
			return true
		}
		return false
	}
	return true
}
