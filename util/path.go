package util

import (
	"errors"
	"io/ioutil"
	"path/filepath"
)

var ErrNotGitRepository = errors.New("not git repository")

// FindGitRoot はpath で指定したリポジトリのルートディレクトリを返す
func FindGitRoot(path string)(string,error){
	files,err := ioutil.ReadDir(path)
	if err != nil{
		return "",err
	}
	for _,file := range files{
		if file.IsDir() && file.Name() == ".git"{
			return path,nil
		}
	}
	abs,err := filepath.Abs(path)
	if err != nil{
		return "",err
	}
	if abs == "/"{
		return "",ErrNotGitRepository
	}

	return FindGitRoot(filepath.Join(path,"."))
}
