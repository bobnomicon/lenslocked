package main

import (
	"fmt"
	"net/http"
)

func homeHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")
	fmt.Fprint(w, "<h1>Welcome to my awesome site!</h1>")
}

func contactHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")
	fmt.Fprint(w, "To get in touch, please send an email to <a href=\"mailto:miller.robert.john@gmail.com\">miller.robert.john@gmail.com</a>.")
}

func main() {
	http.HandleFunc("/", homeHandler)
	http.HandleFunc("/contact", contactHandler)

	fmt.Println("Starting the server on :3000...")
	
	http.ListenAndServe(":3000", nil)
}