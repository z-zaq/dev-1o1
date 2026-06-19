package main

import (
	"html/template"
	"log"
	"net/http"
)

func main() {
	http.Handle("/static/",
		http.StripPrefix("/static/",
			http.FileServer(http.Dir("static"))))
	http.HandleFunc("/", HomeHandler)
	http.HandleFunc("/about", AboutHandler)
	http.HandleFunc("/contact", ContactHandler)

	log.Println("Server running on http://localhost:8081")
	err := http.ListenAndServe(":8081", nil)
	if err != nil {
		log.Fatal(err)
	}
}
func HomeHandler(w http.ResponseWriter, r *http.Request) {
	tmpl, err := template.ParseFiles("templates/home.html")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	tmpl.Execute(w, nil)
}
func AboutHandler(w http.ResponseWriter, r *http.Request) {
	tmpl, err := template.ParseFiles("templates/about.html")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	tmpl.Execute(w, nil)
}
func ContactHandler(w http.ResponseWriter, r *http.Request) {
	tmpl, err := template.ParseFiles("templates/contact.html")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	tmpl.Execute(w, nil)
}
