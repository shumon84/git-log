package store

import (
	"encoding/hex"
	"testing"
)

func TestClient_GetObject(t *testing.T) {
	client, err := NewClient("../testrepo")
	if err != nil{
		t.Fatal(err)
	}
	hashString := "1d1e8977d6596d11345d512c88404e56fb8e46b8"
	hash,err := hex.DecodeString(hashString)
	if err != nil{
		t.Fatal(err)
	}
	obj, err := client.GetObject(hash)
	if err != nil{
		t.Fatal(err)
	}
	t.Log(string(obj.Data))
}
