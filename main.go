package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/gorilla/mux"
	_ "github.com/lib/pq"
	"github.com/mrojasb2000/go-fullystacked-api/gen"
	"github.com/mrojasb2000/go-fullystacked-api/helpers"
	"github.com/mrojasb2000/go-fullystacked-api/models"
)

var (
	dbQuery *gen.Queries
)

func initDatabase() {
	dbURI := fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=disable",
		GetAsString("DB_USER", "postgres"),
		GetAsString("DB_PASSWORD", "postgres"),
		GetAsString("DB_HOST", "localhost"),
		GetAsInt("DB_PORT", 5432),
		GetAsString("DB_NAME", "db"),
	)

	// Open the database
	db, err := sql.Open("postgres", dbURI)
	if err != nil {
		panic(err)
	}

	// Connectivity check
	if err := db.Ping(); err != nil {
		log.Fatalln("Error from database ping:", err)
	}

	// Create the store
	dbQuery = gen.New(db)

	ctx := context.Background()

	createTestingUserDb(ctx)

	if err != nil {
		os.Exit(1)
	}
}

func createTestingUserDb(ctx context.Context) {
	name := "Jhon Doe"
	userName := "jhondeo@company.com"
	password := "password"
	//has the user been created
	u, _ := dbQuery.GetUserByName(ctx, userName)

	if u.UserName == userName {
		log.Println("User exists...")
		return
	}
	log.Println("Creating user testing...")
	hashPwd, _ := helpers.HashPassword(password)
	_, err := dbQuery.CreateUsers(ctx, gen.CreateUsersParams{
		UserName:     userName,
		PassWordHash: hashPwd,
		Name:         name,
	})
	if err != nil {
		log.Println("error getting user testing ", err)
		os.Exit(1)
	}
}

func validateUser(username string, password string) bool {
	ctx := context.Background()
	u, _ := dbQuery.GetUserByName(ctx, username)

	if u.UserName != username {
		return false
	}

	return helpers.CheckPasswordHash(password, u.PassWordHash)
}

func basicMiddleware(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Println("Middleware called on", r.URL.Path)
		// do stuff
		h.ServeHTTP(w, r)
	})
}

func hanlderUserAuth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	var user models.User
	_ = json.NewDecoder(r.Body).Decode(&user)

	if validateUser(user.Name, user.Password) {
		type respose struct {
			Message string
		}
		resp := respose{
			Message: "User Authenticated",
		}
		json.NewEncoder(w).Encode(resp)
	}
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
	router.Use(basicMiddleware)
	srv := http.Server{
		Addr: ":" + port,
		// Addr optionally
		// specifies the listen address for the
		// server in the form of "host:port".
		Handler: router,
	}

	router.HandleFunc("/", hanlderUserAuth).Methods(http.MethodPost)

	go func() {
		initDatabase()
	}()

	log.Println("Starting on port: ", port)
	err := srv.ListenAndServe()
	if err != nil {
		log.Fatalln("Counldn't Listen and Serve :", err)
	}
	fmt.Println("Server is running on port: ", port)
}
