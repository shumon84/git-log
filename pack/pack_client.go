package pack

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/shumon84/git-log/sha"
	"github.com/shumon84/git-log/util"
)

type Client struct {
	idxFileList []string
	idxMap      map[string]*Idx
}

func NewPackClient(path string) (*Client, error) {
	idxFileList, err := getIdxFileList(path)
	if err != nil {
		return nil, err
	}
	return &Client{
		idxFileList: idxFileList,
		idxMap:      map[string]*Idx{},
	}, nil
}

func getIdxFileList(path string) ([]string, error) {
	dir, err := util.FindGitRoot(path)
	if err != nil {
		return nil, err
	}

	packFiles, err := filepath.Glob(filepath.Join(dir, ".git", "objects", "pack", "*.idx"))
	if err != nil {
		return nil, err
	}
	return packFiles, nil
}

func (c *Client) GetIdx(idxFile string) (*Idx, error) {
	indexInfo, ok := c.idxMap[idxFile]
	if ok {
		return indexInfo, nil
	}
	file, err := os.Open(idxFile)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	indexInfo, err = ReadIdxFile(file)
	if err != nil {
		return nil, err
	}
	c.idxMap[idxFile] = indexInfo

	return indexInfo, nil
}

func (c *Client) GetEntry(hash sha.SHA1) (*Entry, error) {
	idxEntry, idxFileName, err := c.GetIdxEntry(hash)
	if err != nil {
		return nil, err
	}
	packFile, err := c.openPackFile(idxFileName)
	if err != nil {
		return nil, err
	}
	defer packFile.Close()

	entry, err := readEntry(packFile, idxEntry)
	if err != nil {
		return nil, err
	}
	entry.Hash = idxEntry.Hash
	return entry, nil
}

func (c *Client) GetIdxEntry(hash sha.SHA1) (*IdxEntry, string, error) {
	for _, idxFileName := range c.idxFileList {
		idx, err := c.GetIdx(idxFileName)
		if err != nil {
			return nil, "", err
		}
		idxEntry, err := idx.Find(hash)
		if err != nil {
			if err == ErrNotFoundIdxEntry {
				continue
			} else {
				return nil, "", err
			}
		}
		return idxEntry, idxFileName, nil
	}
	return nil, "", ErrNotFoundIdxEntry
}

func (c *Client) openPackFile(idxFileName string) (*os.File, error) {
	packFileName := strings.TrimSuffix(idxFileName, ".idx") + ".pack"
	packFile, err := os.Open(packFileName)
	if err != nil {
		return nil, err
	}
	return packFile, nil
}
