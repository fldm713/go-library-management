package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

	_ "github.com/go-sql-driver/mysql"
	"github.com/gorilla/mux"
)

const (
	API_PATH = "/api/v1/books"
)

type library struct {
	dbHost, dbPassword, dbName string
}

type Book struct {
	Id, Name, Isbn string
}

func main() {
	dbHost := os.Getenv("DB_HOST")
	if dbHost == "" {
		dbHost = "localhost:3306"
	}
	dbPassword := os.Getenv("DB_PASS")
	if dbPassword == "" {
		dbPassword = "mysqlpassword"
	}

	apiPath := os.Getenv("API_PATH")
	if apiPath == "" {
		apiPath = API_PATH
	}

	dbName := os.Getenv("DB_NAME")
	if dbName == "" {
		dbName = "library"
	}

	l := library{
		dbHost: dbHost,
		dbPassword: dbPassword,
		dbName: dbName,
	}
	r := mux.NewRouter()
	r.HandleFunc(apiPath, l.getBooks).Methods(http.MethodGet)
	r.HandleFunc(apiPath, l.postBook).Methods(http.MethodPost)
	http.ListenAndServe(":8080", r)
}

func (l library) postBook(w http.ResponseWriter, r *http.Request) {
	log.Println("post was called")
	// read request into an instance of a book
	book := Book{}
	json.NewDecoder(r.Body).Decode(&book)
	// open connection
	db := l.openConnection()
	// write the data
	insertQuery, err := db.Prepare("insert into books values (?, ?, ?)")
	if err != nil {
		log.Fatalf("preparing the db query: %v", err)
	}
	tx, err := db.Begin()
	if err != nil {
		log.Fatalf("while beginning transaction: %v", err)
	}

	_, err = tx.Stmt(insertQuery).Exec(book.Id, book.Name, book.Isbn)
	if err != nil {
		log.Fatalf("execing the insert query: %v", err)
	}

	err = tx.Commit()
	if err != nil {
		log.Fatalf("execing the commit query: %v", err)
	}
	// close connection
	l.closeConnection(db)
}

func (l library) getBooks(w http.ResponseWriter, r *http.Request) {
	// open connection
	// read all book
	// close connection
	log.Println("getbooks was called")
	db := l.openConnection()
	rows, err := db.Query("select * from books")
	if err != nil {
		log.Fatalf("querying the books failed: %v", err)
	}

	books := []Book{}
	for rows.Next() {
		var id, name, isbn string
		err := rows.Scan(&id, &name, &isbn)
		if err != nil {
			log.Fatalf("scanning the books failed: %v", err)
		}
		aBook := Book{Id: id, Name: name, Isbn: isbn}
		books = append(books, aBook)
	}

	json.NewEncoder(w).Encode(books)
	l.closeConnection(db)
}

func (l library) openConnection() *sql.DB {
	db, err := sql.Open("mysql", fmt.Sprintf("%s:%s@(%s)/%s", "root", l.dbPassword, l.dbHost, l.dbName))
	if err != nil {
		log.Fatalf("Opening the connection to database %s failed: %v", l.dbName, err)
	}
	return db
}

func (l library) closeConnection(db *sql.DB) {
	err := db.Close()
	if err != nil {
		log.Fatalf("Closing the connection to database %s failed: %v", l.dbName, err)
	}
}