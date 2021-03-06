package main

import (
	"log"
	"net/http"
	"text/template"
)

var homeTemplate = template.Must(template.ParseFiles("home.html"))

var (
	configuration Configuration
)

func serveHome(w http.ResponseWriter, r *http.Request) {
	log.Println(r.URL)
	if r.URL.Path != "/" {
		http.Error(w, "Not found", 404)
		return
	}
	if r.Method != "GET" {
		http.Error(w, "Method not allowed", 405)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	homeTemplate.Execute(w, r.Host)
}

func main() {
	connectedClients = make(map[string]*Client)
	e := readConfig()
	if e != nil {
		log.Fatalf("Fail in configuration", e)
	}
	http.HandleFunc("/", serveHome)
	http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		serveWs(w, r)
	})
	log.Println("Running server at ", configuration.Port)
	err := http.ListenAndServe(configuration.Port, nil)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}

}
