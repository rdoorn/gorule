package main

import (
	"log"
	"net/http"

	"github.com/rdoorn/gorule"
)

func main() {
	req, _ := http.NewRequest("GET", "http://www.tweakers.net", nil)
	script := ` request.url.path = "/about" `
	err := gorule.Parse(map[string]interface{}{"request": req}, []byte(script))
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("request path modified: %s", req.URL.Path)
}
