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

	router.HandleFunc("/", handlerGetHelloWorld)

	err := srv.ListenAndServe()
	if err != nil {
		log.Fatalln("Couldn't Listen and Server: ", err)
	}
}
