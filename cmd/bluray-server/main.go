package main

import (
	"fmt"
	"net/http"
	"os"

	_ "github.com/sjpotter/bluray-http-server/pkg/pkg/handlers"
)

func main() {
	server := &http.Server{
		Addr: ":8080",
	}

	err := server.ListenAndServe()
	if err != nil {
		fmt.Printf("ListenAndServe Failed: %v\n", err)
		os.Exit(1)
	}
}
