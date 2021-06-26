package store

import (
	"bytes"
	"compress/zlib"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/shumon84/git-log/pack/instruction"

	"github.com/shumon84/git-log/object"
	"github.com/shumon84/git-log/pack"
	"github.com/shumon84/git-log/sha"
	"github.com/shumon84/git-log/util"
)

type Client struct {
	objectsDir string
	PackClient *pack.Client
}

func NewClient(path string) (*Client, error) {
	rootDir, err := util.FindGitRoot(path)
	if err != nil {
		return nil, err
	}
	packClient, err := pack.NewPackClient(path)
	if err != nil {
		return nil, err
	}
	return &Client{
		objectsDir: filepath.Join(rootDir, ".git", "objects"),
		PackClient: packClient,
	}, err
}

func (c *Client) getObjectFromLooseObject(hash sha.SHA1) (*object.Object, error) {
	hashString := hash.String()
	objectPath := filepath.Join(c.objectsDir, hashString[:2], hashString[2:])

	objectFile, err := os.Open(objectPath)
	if err != nil {
		return nil, err
	}
	defer objectFile.Close()
	zr, err := zlib.NewReader(objectFile)
	if err != nil {
		return nil, err
	}

	obj, err := object.ReadObject(zr)
	if err != nil {
		return nil, err
	}

	return obj, nil
}

func (c *Client) getObjectFromPack(hash sha.SHA1) (*object.Object, error) {
	entry, err := c.PackClient.GetEntry(hash)
	if err != nil {
		return nil, err
	}

	if entry.IsDeltaObject() {
		appliedObject, err := c.applyDelta(entry)
		if err != nil {
			return nil, err
		}
		return appliedObject, nil
	}
	looseObject, err := entry.ToObject()
	if err != nil {
		return nil, err
	}
	return looseObject, nil
}

func (c *Client) GetObject(hash sha.SHA1) (*object.Object, error) {
	obj, err := c.getObjectFromLooseObject(hash)
	if err == nil {
		return obj, nil
	}
	if !os.IsNotExist(err) {
		return nil, err
	}

	// .git/objects の中に存在しなかった場合、pack ファイルから探す
	obj, err = c.getObjectFromPack(hash)
	if err != nil {
		return nil, err
	}

	return obj, nil
}

func (c *Client) applyDelta(deltaEntry *pack.Entry) (*object.Object, error) {
	deltaObject, baseObject, err := c.getDeltaAndBaseObject(deltaEntry)
	if err != nil {
		return nil, err
	}

	runner, err := instruction.NewRunner(deltaObject, baseObject)
	if err != nil {
		return nil, err
	}
	result, err := runner.Run()
	if err != nil {
		return nil, err
	}

	return result, nil
}

func (c *Client) getDeltaAndBaseObject(deltaEntry *pack.Entry) ([]byte, *object.Object, error) {
	var deltaData []byte
	var baseObjectHash sha.SHA1
	switch deltaEntry.Type {
	case pack.OBJ_OFS_DELTA:
		negativeOffset, n, err := util.DecodeOffsetEncoding(bytes.NewBuffer(deltaEntry.Data))
		if err != nil {
			return nil, nil, err
		}
		_, idxFileName, err := c.PackClient.GetIdxEntry(deltaEntry.Hash)
		if err != nil {
			return nil, nil, err
		}
		idx, err := c.PackClient.GetIdx(idxFileName)
		if err != nil {
			return nil, nil, err
		}
		baseIdxEntry, err := idx.FindByOffset(deltaEntry.Offset - int64(negativeOffset))
		if err != nil {
			return nil, nil, err
		}
		baseObjectHash = baseIdxEntry.Hash
		deltaData = deltaEntry.Data[n:]
	case pack.OBJ_REF_DELTA:
		baseObjectHash = deltaEntry.Data[:20]
		deltaData = deltaEntry.Data[20:]
	default:
		return nil, nil, ErrNotDeltaObject
	}
	baseObject, err := c.GetObject(baseObjectHash)
	if err != nil {
		return nil, nil, err
	}

	deltaDataZlibReader, err := zlib.NewReader(bytes.NewBuffer(deltaData))
	if err != nil {
		return nil, nil, err
	}
	deltaObject, err := ioutil.ReadAll(deltaDataZlibReader)
	if err != nil {
		return nil, nil, err
	}

	return deltaObject, baseObject, nil
}

type WalkFunc func(object *object.Commit) error

func (c *Client) WalkHistory(hash sha.SHA1, walkFunc WalkFunc) error {
	// Git のコミットグラフは、際限なく深くなるものの、幅はそれほどなケースがほとんどなので、ここでは BFS で実装してる
	// でも本物の git log は人間が出力を眺めやすいように DFS で実装されてる
	ancestors := []sha.SHA1{hash}
	cycleCheck := map[string]struct{}{}
	for len(ancestors) > 0 {
		currentHash := ancestors[0]
		if _, ok := cycleCheck[string(currentHash)]; ok {
			ancestors = ancestors[1:]
			continue
		}
		cycleCheck[string(currentHash)] = struct{}{}

		obj, err := c.GetObject(currentHash)
		if err != nil {
			return err
		}

		current, err := object.NewCommit(obj)
		if err != nil {
			return err
		}

		if err := walkFunc(current); err != nil {
			return err
		}

		// ここパフォーマンス悪いので、本当はちゃんと Queue を実装してね
		ancestors = append(ancestors[1:], current.Parents...)
	}
	return nil
}
