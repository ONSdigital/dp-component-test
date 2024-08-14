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
				<label for="example-input">This is a test label</label>
				<input id='example-input' class='example-input' value=''>
				<button class="example-button" onclick="changeValue()">Click me</button>
				<script>
					function changeValue() {
						document.getElementById("example-input").value = "clicked";
					}
				</script>
			</body>
		</html>`
	fmt.Fprint(w, htmlPage)
}

func AccessibilityExclusionHandler(w http.ResponseWriter, _ *http.Request) {
	htmlPage := `
		<!doctype html>
		<html lang=en>
			<head>
				<meta charset=utf-8>
				<title>blah</title>
			</head>
			<body>
				<p class='example-paragraph'>Example web page</p>
				<label for="test">This is a test label</label>
				<input id="test" class='example-input' value='test value'>
				<!-- This is an accessibility failure -->
				<image src="/example.png" />
			</body>
		</html>`
	fmt.Fprint(w, htmlPage)
}

func newRouter() http.Handler {
	router := mux.NewRouter().StrictSlash(true)
	router.HandleFunc("/example", ExampleHandler1).Methods("GET")
	router.HandleFunc("/example-accessibility-exclusion", AccessibilityExclusionHandler).Methods("GET")
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
