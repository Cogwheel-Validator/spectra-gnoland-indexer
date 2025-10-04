package main

import (
	"encoding/base64"
	"fmt"
	"log"
)

func main() {
	fmt.Println("Enter base64 to encode to base64url:  ")
	var input string
	fmt.Scanln(&input)
	base64Decoded, err := base64.StdEncoding.DecodeString(input)
	if err != nil {
		log.Fatal(err)
	}
	base64url := base64.URLEncoding.EncodeToString(base64Decoded)
	fmt.Println("Base64url: ", base64url)
}
