package main

import (
	"log"

	"github.com/lisp-ceo/dlog/internal/server"
)

func main() {
	srv := server.NewHTTPServer(":8888")
	log.Fatal(srv.ListenAndServe())
}
