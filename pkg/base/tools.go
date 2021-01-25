package base

import (
	"fmt"
	"net/url"
	"os"
	"path"
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
	case "date":
	case "time":
	case "datetime":
	case "timestamp":
	default:
		return false
	}
	return true
}

func HasHttpPrefix(path string) bool {
	return strings.HasPrefix(path, "https://") || strings.HasPrefix(path, "http://")
}

func ExtractFilename(uri string) (local bool, filename string, err error) {
	if !HasHttpPrefix(uri) {
		local, filename, err = true, uri, nil
		return
	}

	local = false
	base := path.Base(uri)
	if index := strings.Index(base, "?"); index != -1 {
		filename, err = url.QueryUnescape(base[:index])
	} else {
		filename, err = url.QueryUnescape(base)
	}
	return
}

func FormatFilePath(filepath string) (path string, err error) {
	local, path, err := ExtractFilename(filepath)
	if local || err != nil {
		return
	}
	path = fmt.Sprintf("http(s)://**/%s", path)
	return
}
