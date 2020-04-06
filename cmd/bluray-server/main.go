package main

import (
	"flag"
	"fmt"
	"net/http"
	"os"

	_ "github.com/sjpotter/bluray-http-server/pkg/pkg/handlers"
)

var (
	port = flag.Int("port", 8080, "port for http server to listen to")
)

func main() {
	flag.Parse()

	server := &http.Server{
		Addr: fmt.Sprintf(":%v", *port),
	}

	err := server.ListenAndServe()
	if err != nil {
		fmt.Printf("ListenAndServe Failed: %v\n", err)
		os.Exit(1)
	}
}
