package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/gorilla/mux"
	_ "github.com/lib/pq"
	"github.com/ory/dockertest/v3"
)

var (
	DD *sql.DB
)

func main() {
	// Инициализация роутера Gin
	router := mux.NewRouter()
	// Запуск контейнера PostgreSQL
	pool, err := dockertest.NewPool("")
	if err != nil {
		log.Fatalf("Could not connect to Docker: %s", err)
	}

	options := &dockertest.RunOptions{
		Repository: "postgres",
		Tag:        "latest",
		Env: []string{
			"POSTGRES_DB=productdb",
			"POSTGRES_USER=postgres",
			"POSTGRES_PASSWORD=1234",
		},
	}

	resource, err := pool.RunWithOptions(options)
	if err != nil {
		log.Fatalf("Could not start PostgreSQL container: %s", err)
	}
	defer pool.Purge(resource)

	// Ожидание готовности контейнера PostgreSQL
	if err := pool.Retry(func() error {
		db, err := sql.Open("postgres", getDBConnectionString())
		DD = db
		if err != nil {
			log.Println("Waiting for PostgreSQL to be ready...")
			return err
		}
		return db.Ping()
	}); err != nil {
		log.Fatalf("Failed to connect to PostgreSQL: %s", err)
	}
	router.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, "<html><head><title>My Golang Web App</title></head><body><h1>Hello, Golang Web App!</h1></body></html>")
	}).Methods("GET")
	router.HandleFunc("/products", GetProducts).Methods("GET") //Получить информацию о всех продуктах

	// Подключение к базе данных PostgreSQL
	db, err := sql.Open("postgres", getDBConnectionString())
	DD = db
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// Запуск сервера
	port := ":8080"
	log.Fatal(http.ListenAndServe(port, router))
	fmt.Printf("Server is running on port %s\n", port)
}

func getDBConnectionString() string {
	host := os.Getenv("DB_HOST")
	port := os.Getenv("DB_PORT")
	user := os.Getenv("DB_USER")
	password := os.Getenv("DB_PASSWORD")
	dbname := os.Getenv("DB_NAME")

	return fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable", host, port, user, password, dbname)
}
func GetProducts(w http.ResponseWriter, r *http.Request) {
	fmt.Println("КАКА2")
	prods := GetData()
	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(prods)
}
func GetData() []Product {
	rows, err := DD.Query("SELECT * FROM products")
	fmt.Println("КАКА2.1")
	if err != nil {
		fmt.Println("КАКА2.1.4")
		log.Fatal(err)
	}
	fmt.Println("КАКА2.1.1")
	defer rows.Close()
	var prod []Product
	// Обработка результатов запроса
	for rows.Next() {
		var id, naming, description string
		var weight float64
		err := rows.Scan(&id, &naming, &weight, &description)
		if err != nil {
			fmt.Println("КАКА2.1.2")
			panic(err)
		}
		prod = append(prod, Product{ID: id, Naming: naming, Weight: weight, Description: description})
	}
	fmt.Println("КАКА2.2")

	return prod
}

type Product struct {
	ID          string  `json:"item_id"`
	Naming      string  `json:"name"`
	Weight      float64 `json:"weight"`
	Description string  `json:"description"`
}
