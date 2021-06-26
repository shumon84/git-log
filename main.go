package main

import (
	"encoding/hex"
	"fmt"
	"log"
	"os"

	"github.com/shumon84/git-log/store"

	"github.com/shumon84/git-log/object"
)

func main() {
	hashString := os.Args[1]
	hash, err := hex.DecodeString(hashString)
	if err != nil {
		log.Fatal(err)
	}

	client, err := store.NewClient(".")
	if err != nil {
		log.Fatal(err)
	}

	err = client.WalkHistory(hash, func(object *object.Commit) error {
		fmt.Println(object)
		fmt.Println("")
		return nil
	})
	if err != nil {
		log.Fatal(err)
	}
}
