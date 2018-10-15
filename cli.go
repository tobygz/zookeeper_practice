package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
)

func batchHttp(a int, b int) {
	urlstr := fmt.Sprintf("http://127.0.0.1:2101/?a=%d&b=%d", a, b)
	resp, err := http.Get(urlstr)
	if err != nil {
		panic(err)
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		panic(err)
	}
	log.Println(string(body))
	resp.Body.Close()
}

func main() {

	for i := 0; i < 1024; i++ {
		batchHttp(i, i)
	}
}
