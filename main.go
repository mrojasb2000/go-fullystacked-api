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
	"github.com/gorilla/sessions"
	_ "github.com/lib/pq"
	"github.com/mrojasb2000/go-fullystacked-api/gen"
	"github.com/mrojasb2000/go-fullystacked-api/helpers"
	"github.com/mrojasb2000/go-fullystacked-api/models"

	"github.com/go-redis/redis/v8"
	rstore "github.com/rbcervilla/redisstore/v8"
)

var (
	dbQuery *gen.Queries

	store = sessions.NewCookieStore([]byte("forDemo"))
)

// securityMiddleware is middleware to make sure all request has a
// valid session and authenticated
func securityMiddleware(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Println("Security middleware called on", r.URL.Path)

		// first of all request MUST have a valid session
		if sessionValid(w, r) {
			// login path will be let through
			if r.URL.Path == "/login" {
				h.ServeHTTP(w, r)
				return
			}
		}

		// if it does have a valid session make sure it has been authentication
		if hasBeenAuthenticated(w, r) {
			log.Println("Has been authenticated?")
			// do stuff
			h.ServeHTTP(w, r)
		}
		// do stuff
		h.ServeHTTP(w, r)
		// otherwise it will to be redirected to /login
		storeAuthenticated(w, r, false)
		//http.Redirect(w, r, "/login", http.StatusTemporaryRedirect)
	})
}

// Check whether the session is a valid session
func sessionValid(w http.ResponseWriter, r *http.Request) bool {
	session, _ := store.Get(r, "session_token")
	return !session.IsNew
}

// hasBeenAuthenticated checks whether the session contain the flag to indicate
// that the session has gone through authentication process
func hasBeenAuthenticated(w http.ResponseWriter, r *http.Request) bool {
	session, _ := store.Get(r, "session_token")

	isAuthenticated := session.Values["authenticated"]

	if isAuthenticated == nil {
		fmt.Println("User not authenticated")
		return false
	}
	return isAuthenticated.(bool)
}

func loginHandler(w http.ResponseWriter, r *http.Request) {
	var user models.User
	_ = json.NewDecoder(r.Body).Decode(&user)

	if validateUser(user.Name, user.Password) {
		storeAuthenticated(w, r, true)
		renderResponseJson(w, r, "User Authenticated")
		log.Println("User Authenticated")
		return
	}
	renderResponseJson(w, r, "User Not Authenticated")
}

// storeAuthenticated to store authenticated value
func storeAuthenticated(w http.ResponseWriter, r *http.Request, v bool) {
	session, _ := store.Get(r, "session_token")

	session.Values["authenticated"] = v
	err := session.Save(r, w)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
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

func renderResponseJson(w http.ResponseWriter, r *http.Request, message string) {
	w.Header().Set("Content-Type", "application/json")
	type respose struct {
		Message string
	}

	json.NewEncoder(w).Encode(respose{
		Message: message,
	})
}

func logoutHanlder(w http.ResponseWriter, r *http.Request) {
	if hasBeenAuthenticated(w, r) {
		session, _ := store.Get(r, "session_token")
		session.Options.MaxAge = -1
		err := session.Save(r, w)
		if err != nil {
			log.Println("failed to delete session ", err)
		}
	}
	//http.Redirect(w, r, "/login", http.StatusTemporaryRedirect)
	log.Println("Redirect to login page")
}

func main() {
	// initialize databases and local store
	go func() {
		initDatabase()
		initRedis()
	}()

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
	router.Use(securityMiddleware)
	srv := http.Server{
		Addr: ":" + port,
		// Addr optionally
		// specifies the listen address for the
		// server in the form of "host:port".
		Handler: router,
	}

	router.HandleFunc("/login", loginHandler).Methods(http.MethodPost)
	router.HandleFunc("/logout", logoutHanlder).Methods(http.MethodPost)

	log.Println("Starting on port: ", port)
	err := srv.ListenAndServe()
	if err != nil {
		log.Fatalln("Counldn't Listen and Serve :", err)
	}
	fmt.Println("Server is running on port: ", port)
}

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

	log.Println("Database initialized")
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

func initRedis() {
	var err error

	client := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
	})

	store, err := rstore.NewRedisStore(context.Background(), client)
	if err != nil {
		log.Fatal("failed to create redis store: ", err)
		os.Exit(1)
	}

	store.KeyPrefix("session_token")

	log.Println("Store initialized")
}
