package store

import (
	"compress/zlib"
	"github.com/shumon84/git-log/object"
	"github.com/shumon84/git-log/sha"
	"github.com/shumon84/git-log/util"
	"os"
	"path/filepath"
)

type Client struct {
	objectDir string
}

func NewClient(path string) (*Client,error){
	rootDir, err := util.FindGitRoot(path)
	if err != nil{
		return nil,err
	}
	return &Client{
		objectDir: filepath.Join(rootDir,".git","objects"),
	},nil
}

// GetObject は hash で指定した object を返す
func (c *Client)GetObject(hash sha.SHA1)(*object.Object,error){
	hashString := hash.String()
	objectPath := filepath.Join(c.objectDir,hashString[:2],hashString[2:])

	objectFile,err := os.Open(objectPath)
	if err != nil{
		return nil,err
	}
	defer objectFile.Close()

	zr, err := zlib.NewReader(objectFile)
	if err != nil{
		return nil,err
	}

	obj,err := object.ReadObject(zr)
	if err !=nil{
		return nil,err
	}

	return obj,nil
}

type WalkFunc func(commit *object.Commit)error

// WalkHistory は hash で指定したコミットから、履歴をさかのぼって、それぞれのコミットに walkFunc を適用する
func (c *Client) WalkHistory(hash sha.SHA1,walkFunc WalkFunc)error{
	ancestors := []sha.SHA1{hash}
	cycleCheck := map[string]struct{}{}

	// BFS
	for len(ancestors) > 0{
		currentHash := ancestors[0]
		if _,ok := cycleCheck[string(currentHash)];ok{
			ancestors = ancestors[1:]
			continue
		}
		cycleCheck[string(currentHash)] = struct{}{}

		obj,err := c.GetObject(currentHash)
		if err != nil{
			return err
		}

		current,err := object.NewCommit(obj)
		if err != nil{
			return err
		}

		if err := walkFunc(current); err != nil{
			return err
		}

		ancestors = append(ancestors[1:],current.Parents...)
	}

	return nil
}
