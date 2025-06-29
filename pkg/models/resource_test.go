package models

import (
	"encoding/json"
	"fmt"
	"log"
	"net/url"
	"testing"
)

func Test(t *testing.T) {
	parse, err := url.Parse("https://qq@www.jetbrains.com/")
	if err != nil {
		log.Fatal(err)
	}
	m := Resource(*parse)
	marshal, err := json.Marshal(&m)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(string(marshal))
	var u *Resource
	err = json.Unmarshal(marshal, &u)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(parse.String())
	fmt.Println(u)
}
