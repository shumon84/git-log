package util

import (
	"errors"
	"io/ioutil"
	"path/filepath"
)

var (
	ErrNotGitRepository = errors.New("not git repository")
)

func FindGitRoot(path string) (string, error) {
	files, err := ioutil.ReadDir(path)
	if err != nil {
		return "", err
	}
	for _, file := range files {
		if file.IsDir() && file.Name() == ".git" {
			return path, nil
		}
	}
	if abs, err := filepath.Abs(path); err != nil {
		return "", err
	} else if abs == "/" {
		return "", ErrNotGitRepository
	}
	return FindGitRoot(filepath.Join(path, ".."))
}
