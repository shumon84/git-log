package object

import (
	"bufio"
	"bytes"
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"io"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/shumon84/git-log/sha"
)

type Commit struct {
	Hash      sha.SHA1
	Size      int
	Tree      sha.SHA1
	Parents   []sha.SHA1
	Author    Sign
	Committer Sign
	Message   string
}

func (c Commit) String() string {
	str := ""
	str += fmt.Sprintln("Commit   ", c.Hash)
	str += fmt.Sprintln("Tree     ", c.Tree)
	for _, parent := range c.Parents {
		str += fmt.Sprintln("Parent   ", parent)
	}
	str += fmt.Sprintln("Author   ", c.Author)
	str += fmt.Sprintln("Committer", c.Committer)
	str += fmt.Sprint(c.Message)
	return str
}

type Sign struct {
	Name      string
	Email     string
	Timestamp time.Time
}

func (s Sign) String() string {
	return fmt.Sprintf("%s %s %s", s.Name, s.Email, s.Timestamp.String())
}

var (
	emailRegexpString     = "([a-zA-Z0-9_.+-]+@([a-zA-Z0-9][a-zA-Z0-9-]*[a-zA-Z0-9]*\\.)+[a-zA-Z]{2,})"
	timestampRegexpString = "([1-9][0-9]* \\+[0-9]{4})"
	sha1Regexp            = regexp.MustCompile("[0-9a-f]{20}")
	signRegexp            = regexp.MustCompile("^[^<]* <" + emailRegexpString + "> " + timestampRegexpString + "$")
)

// NewCommit は *Object を *Commit に変換して返す
func NewCommit(o *Object) (*Commit, error) {
	if o.Type != CommitObject {
		return nil, ErrNotCommitObject
	}

	checkSum := sha1.New()
	b := bytes.NewBuffer(o.Data)
	tr := io.TeeReader(b, checkSum)

	checkSum.Write(o.Header())

	commit := &Commit{
		Size: o.Size,
	}

	scanner := bufio.NewScanner(tr)
	for scanner.Scan() {
		text := scanner.Text()
		splitText := strings.SplitN(text, " ", 2)
		if len(splitText) != 2 {
			break
		}
		lineType := splitText[0]
		data := splitText[1]

		switch lineType {
		case "tree":
			tree, err := readHash(data)
			if err != nil {
				return nil, err
			}
			commit.Tree = tree
		case "parent":
			parent, err := readHash(data)
			if err != nil {
				return nil, err
			}
			commit.Parents = append(commit.Parents, parent)
		case "author":
			author, err := readSign(data)
			if err != nil {
				return nil, err
			}
			commit.Author = author
		case "committer":
			committer, err := readSign(data)
			if err != nil {
				return nil, err
			}
			commit.Committer = committer
		}
	}

	message := make([]string, 0)
	for scanner.Scan() {
		message = append(message, scanner.Text())
	}
	commit.Message = strings.Join(message, "\n")

	hash := checkSum.Sum(nil)
	if string(o.Hash) != string(hash) {
		return nil, ErrInvalidCommitObject
	}
	commit.Hash = hash
	return commit, nil
}

func readHash(hashString string) (sha.SHA1, error) {
	if ok := sha1Regexp.MatchString(hashString); !ok {
		return nil, ErrInvalidCommitObject
	}
	hash := make(sha.SHA1, 20)
	if _, err := hex.Decode(hash, []byte(hashString)); err != nil {
		return nil, fmt.Errorf("%w : %s", ErrInvalidCommitObject, err)
	}
	return hash, nil
}

func readSign(signString string) (Sign, error) {
	if ok := signRegexp.MatchString(signString); !ok {
		return Sign{}, ErrInvalidCommitObject
	}
	sign1 := strings.SplitN(signString, " <", 2)
	name := sign1[0]
	sign2 := strings.SplitN(sign1[1], "> ", 2)
	email := sign2[0]
	sign3 := strings.SplitN(sign2[1], " ", 2)

	unixTime, err := strconv.ParseInt(sign3[0], 10, 64)
	if err != nil {
		return Sign{}, fmt.Errorf("%w : %s", ErrInvalidCommitObject, err)
	}
	var offsetHour, offsetMinute int
	if _, err := fmt.Sscanf(sign3[1], "+%02d%02d", &offsetHour, &offsetMinute); err != nil {
		return Sign{}, fmt.Errorf("%w : %s", ErrInvalidCommitObject, err)
	}
	location := time.FixedZone(" ", 3600*offsetHour+60*offsetMinute)
	timestamp := time.Unix(unixTime, 0).In(location)
	time.Now().String()
	return Sign{
		Name:      name,
		Email:     email,
		Timestamp: timestamp,
	}, nil
}
