package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	_ "github.com/lib/pq"
	"github.com/mrojasb2000/go-fullystacked-api/gen"
)

const DB_DRIVERNAME = "postgres"

func main() {
	dbURI := fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=disable",
		GetAsString("DB_USER", "postgres"),
		GetAsString("DB_PASSWORD", "postgres"),
		GetAsString("DB_HOST", "localhost"),
		GetAsInt("DB_PORT", 5432),
		GetAsString("DB_NAME", "db"),
	)

	// Open the database
	db, err := sql.Open(DB_DRIVERNAME, dbURI)

	if err != nil {
		panic(err)
	}

	// Connectivity check
	if err := db.Ping(); err != nil {
		log.Fatalln("Error from database ping: ", err)
	}

	// Create the store
	st := gen.New(db)

	ctx := context.Background()

	_, err = st.CreateUsers(ctx, gen.CreateUsersParams{
		UserName:     "jhondoe",
		PassWordHash: "hash",
		Name:         "Jhon Doe",
	})

	if err != nil {
		log.Fatalln("Error creating user :", err)
	}

	log.Println("Done!")

	users, _ := st.ListUsers(ctx)

	for _, usr := range users {
		data := fmt.Sprintf("Name: %s, ID: %d", usr.Name, usr.UserID)
		fmt.Println(data)
	}
}
