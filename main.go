package main

import (
	"fmt"
	"log"
	"net/http"
	"time"
)

func handlerGetHelloWorld(w http.ResponseWriter, req *http.Request) {
	fmt.Fprintf(w, "Hello World\n")
	log.Println(req.Method)
	log.Println(req.URL)
	log.Println(req.Header)
	log.Println(req.Body)
}

type foo struct {
}

func (d foo) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	fmt.Fprintf(w, "Hello World Alternative\n")
	log.Println(req.Method)
	log.Println(req.URL)
	log.Println(req.Header)
	log.Println(req.Body)
}

func main() {
	port := "9002"
	router := http.NewServeMux()

	srv := http.Server{
		Addr:           ":" + port,
		Handler:        router,
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   120 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}

	// This is just to show an alternative way to declare a handler
	// by having a struct that implements the ServeHTTP(...) interface
	dummyHandler := foo{}

	router.HandleFunc("/", handlerGetHelloWorld)

	router.Handle("/1", dummyHandler)

	err := srv.ListenAndServe()
	if err != nil {
		log.Fatalln("Couldn't Listen and Server: ", err)
	}
	log.Println("Server is running on port :", port)
}
