package main

import (
	"bytes"
	"encoding/json"
	"io"
	"log"
	"net/http"
)

func main() {
	funSig := "test1"
	repoUrl := "test2"

	input := Input{
		FuncSig: funSig,
		RepoUrl: repoUrl,
	}

	client := &http.Client{}
	resBody, err := json.Marshal(input)
	if err != nil {
		panic(err)
	}

	req, err := http.NewRequest("POST", "http://127.0.0.1:8080/", bytes.NewBuffer(resBody))
	if err != nil {
		panic(err)
	}

	req.SetBasicAuth("cloud", "cloud")

	res, err := client.Do(req)
	if err != nil {
		log.Fatal(err)
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		log.Fatal("http status code not 200")
	}

	resBody, err = io.ReadAll(res.Body)
	if err != nil {
		log.Fatal(err)
	}

	log.Println(string(resBody))
}
