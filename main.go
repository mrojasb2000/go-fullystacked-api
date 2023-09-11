package main

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"

	"github.com/gorilla/mux"
)

func HandlerSlug(w http.ResponseWriter, r *http.Request) {
	slug := mux.Vars(r)["slug"]
	if slug == "" {
		log.Println("Slug not provided")
		return
	}
	log.Println("Got slug, slug")
}

func handlerGetHelloWorld(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Hello World\n")
	log.Println("Request via:\t", r.Method)
	log.Println("Request url:\t", r.URL)
	log.Println("Request Header:\t", r.Header)
	log.Println("Request body:\t", r.Body)
}

func handlerPostEcho(w http.ResponseWriter, r *http.Request) {
	log.Println("Request via:\t", r.Method)
	log.Println("Request url:\t", r.URL)
	log.Println("Request Header:\t", r.Header)

	// We are going to read it into a buffer
	// as the request body is an io.ReadCloser
	// and so we should only read it once.
	body, _ := io.ReadAll(r.Body)
	log.Println("read >t", string(body), "\t<")

	n, err := io.Copy(w, bytes.NewReader(body))
	if err != nil {
		log.Println("Error echoing response: ", err)
	}
	log.Println("Wrote back ", n, " bytes")
}

func main() {
	// Set some flags for easy debugging
	log.SetFlags(log.Lshortfile | log.Ldate | log.Lmicroseconds)

	// Get a port from ENV var or default to 9002
	port := "9002"
	if value, exists := os.LookupEnv("SERVER_PORT"); exists {
		port = value
	}

	// Off the back, we can enforce StrictSlash
	// This is a nice helper function that means
	// When true, if the route is "/foo/",
	// accessing "/foo" will peform a 301
	// redirect to the former and vice versa.
	// In other words, your application will
	// always see the path as specified in the
	// route.
	// When false, if the route path is "/foo"
	// accessing "/foo/" will not math this
	// route and vice versa.
	router := mux.NewRouter().StrictSlash(true)

	srv := http.Server{
		Addr: ":" + port,
		// Addr optionally
		// specifies the listen address for the
		// server in the form of "host:port".
		Handler: router,
	}

	router.HandleFunc("/", handlerGetHelloWorld).Methods(http.MethodGet)
	router.HandleFunc("/", handlerPostEcho).Methods(http.MethodPost)
	router.HandleFunc("/{slug}", HandlerSlug).Methods(http.MethodGet)

	log.Println("Starting on port: ", port)
	err := srv.ListenAndServe()
	if err != nil {
		log.Fatalln("Counldn't Listen and Serve :", err)
	}
	fmt.Println("Server is running on port: ", port)
}
