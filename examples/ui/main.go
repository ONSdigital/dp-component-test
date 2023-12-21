package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/gorilla/mux"
)

func ExampleHandler1(w http.ResponseWriter, _ *http.Request) {
	htmlPage := `
		<!doctype html>
		<html lang=en>
			<head>
				<meta charset=utf-8>
				<title>blah</title>
			</head>
			<body>
				<p class='example-paragraph'>Example web page</p>
				<input class='example-input' value='test value'>
			</body>
		</html>`
	fmt.Fprint(w, htmlPage)
}

func newRouter() http.Handler {
	router := mux.NewRouter().StrictSlash(true)
	router.HandleFunc("/example", ExampleHandler1).Methods("GET")
	return router
}

func NewServer() *http.Server {
	return &http.Server{
		Handler: newRouter(),
	}
}

func main() {
	server := NewServer()
	log.Fatal(http.ListenAndServe(":10000", server.Handler))
}
