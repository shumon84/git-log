package main

import (
	"encoding/hex"
	"fmt"
	"github.com/shumon84/git-log/object"
	"github.com/shumon84/git-log/store"
	"log"
)

func main(){
	//hashString := os.Args[1]
	hashString := "9ca6c931881417a92322c3a380ab46edcee720b9"
	hash,err := hex.DecodeString(hashString)
	if err != nil{
		log.Fatal(err)
	}

	client, err := store.NewClient(".")
	if err != nil{
		log.Fatal()
	}
	if err := client.WalkHistory(hash, func(commit *object.Commit) error {
		fmt.Println(commit)
		fmt.Println("")
		return nil
	});err != nil{
		log.Fatal(err)
	}

}
