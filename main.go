package main

import (
	"fmt"
	"net/http"
)

type Router struct{}

func (router Router) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch r.URL.Path {
	case "/":
		homeHandler(w, r)
	case "/contact":
		contactHandler(w, r)
	case "/faq":
		faqHandler(w, r)
	default:
		http.Error(w, "Page not found", http.StatusNotFound)
	}
}

func pathHandler(w http.ResponseWriter, r *http.Request) {
	switch r.URL.Path {
	case "/":
		homeHandler(w, r)
	case "/contact":
		contactHandler(w, r)
	case "/faq":
		faqHandler(w, r)
	default:
		http.Error(w, "Page not found", http.StatusNotFound)
	}
}

func homeHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")
	fmt.Fprint(w, "<h1>Welcome to my awesome site!</h1>")
}

func contactHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")
	fmt.Fprint(w, "To get in touch, please send an email to <a href=\"mailto:miller.robert.john@gmail.com\">miller.robert.john@gmail.com</a>.")
}

func faqHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	fmt.Fprint(w, `
		<h1>FAQ Page</h1>
		<ul>
			<li><b>Is there a free version?</b> Yes! We offer a free trial for 30 days on any paid plans.</li>
			<li><b>What are your support hours?</b> We have support staff answering emails 24/7, though response times may be a bit slower on weekends.</li>
			<li><b>How do I contact support?</b>Email us - <a href="mailto:support@lenslocked.com">support@lenslocked.com</a></li>
		</ul>
	`)
}

func main() {
	router := http.HandlerFunc(pathHandler)
	fmt.Println("Starting the server on :3000...")
	http.ListenAndServe("localhost:3000", router)
}