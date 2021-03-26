package inventory

import (
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/mitchellh/go-homedir"
)

// GetHomeDir returns the home dir
func GetHomeDir() (string, error) {
	home, err := homedir.Dir()
	if err != nil {
		return "", fmt.Errorf(err.Error())
	}
	return home, nil
}

// ExpandHomePath translate home dir
func ExpandHomePath(path string) string {
	if path != "" && path[:1] == "~" {
		home, err := GetHomeDir()
		if err != nil {
			return ""
		}
		return home + path[1:]
	}
	return path
}

// ReadFile expands home dir
func ReadFile(filename string) ([]byte, error) {
	fn := ExpandHomePath(filename)
	return ioutil.ReadFile(fn)
}

func GetFiles(dirname, filespec string) ([]string, error) {
	files := []string{}
	err := filepath.Walk(dirname, func(dirname2 string, f os.FileInfo, _ error) error {
		if f != nil && !f.IsDir() {
			fname := strings.TrimSuffix(f.Name(), ".yaml")
			if filespec == fname {
				files = append(files, fname)
			} else {
				matched, err := path.Match(filespec, fname)
				if err == nil && matched {
					files = append(files, fname)
				}
			}
		}
		return nil
	})
	return files, err
}
