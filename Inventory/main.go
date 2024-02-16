package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/lib/pq"
	_ "github.com/lib/pq"
)

type Product struct {
	ID       string `json:"item_id"`
	Name     string `json:"name"`
	Quantity int    `json:"quantity"`
	Price    string `json:"price"`
}

var (
	db *sql.DB
)

const (
	host     = "localhost"
	port     = 5432
	user     = "postgres"
	password = "1234"
	dbname   = "inventorydb"
	PortAddr = ":8082"
)

func main() {
	db = ConnectDd()
	defer db.Close()
	fmt.Println("Подключение к PostgreSQL успешно!")
	router := mux.NewRouter()
	router.HandleFunc("/inventory", GETInv).Methods("GET")
	router.HandleFunc("/inventory/{id}", GETInvID).Methods("GET")
	router.HandleFunc("/inventory", CreateInv).Methods("POST")
	router.HandleFunc("/inventory/{id}", UpdInv).Methods("PUT")
	router.HandleFunc("/inventory/{id}", DelInv).Methods("DELETE")

	fmt.Println("Сервер слушате порт " + PortAddr)
	log.Fatal(http.ListenAndServe(PortAddr, router))
}
func GETInv(w http.ResponseWriter, r *http.Request) {
	prods := GetData()
	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(prods)
}
func GETInvID(w http.ResponseWriter, r *http.Request) {
	prod, err := GetDataID(mux.Vars(r)["id"])
	if err != nil {
		if err != sql.ErrNoRows {
			w.WriteHeader(http.StatusNotFound)
			return
		} else {
			log.Fatal(err)
		}
	}
	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(prod)
}
func CreateInv(w http.ResponseWriter, r *http.Request) {
	var prod Product
	_ = json.NewDecoder(r.Body).Decode(&prod)
	_, err := Insert(prod)
	if err != nil {
		if pgErr, ok := err.(*pq.Error); ok && pgErr.Code == "23505" {
			w.WriteHeader(http.StatusNotFound)
			errorResponse := map[string]string{
				"error":   "Product is already exist",
				"message": "The resource with the specified ID already exist.",
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(errorResponse)
			return
		} else {
			log.Fatal(err)
		}
	}
	w.WriteHeader(http.StatusCreated)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(prod)
}
func UpdInv(w http.ResponseWriter, r *http.Request) {
	var prod Product
	_ = json.NewDecoder(r.Body).Decode(&prod)
	_, err := UpdateID(prod)
	if err != nil {
		if err == sql.ErrNoRows {
			w.WriteHeader(http.StatusNotFound)
			errorResponse := map[string]string{
				"error":   "Resource not found",
				"message": "The resource with the specified ID does not exist.",
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(errorResponse)
			return
		} else {
			log.Fatal(err)
		}
	} else {
		w.WriteHeader(http.StatusNoContent)
	}
}
func DelInv(w http.ResponseWriter, r *http.Request) {
	count, err := DeleteID(mux.Vars(r)["id"])
	if err != nil {
		log.Fatal(err)
	}
	if count == 0 {
		w.WriteHeader(http.StatusNotFound)
		errorResponse := map[string]string{
			"error":   "Resource not found",
			"message": "The resource with the specified ID does not exist.",
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(errorResponse)
	} else {
		w.WriteHeader(http.StatusNoContent)
	}
}
func ConnectDd() *sql.DB {
	connStr := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable", host, port, user, password, dbname)
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		fmt.Println(err)
		panic(err)
	}
	return db
}
func Insert(item Product) (int64, error) {
	res, err := db.Exec("INSERT INTO inventory (id, name, quantity, price) VALUES ($1,$2,$3,$4)", item.ID, item.Name, item.Quantity, item.Price)
	if err != nil {
		return -1, err
	}
	rowcount, err := res.RowsAffected()
	if err != nil {
		return rowcount, err
	}
	return rowcount, nil
}
func GetDataID(IDNAME string) (Product, error) { //Обработать ошибку после работы функции
	rows := db.QueryRow("SELECT * FROM inventory WHERE id=$1", IDNAME)
	var prod Product
	// Обработка результатов запроса
	err := rows.Scan(&prod.ID, &prod.Name, &prod.Quantity, &prod.Price)
	if err != nil {
		return Product{}, err
	}
	return prod, nil
}
func GetData() []Product {
	rows, err := db.Query("SELECT * FROM inventory")
	if err != nil {
		panic(err)
	}
	defer rows.Close()
	var prod []Product
	// Обработка результатов запроса
	for rows.Next() {
		var id, name, price string
		var quantity int
		err := rows.Scan(&id, &name, &quantity, &price)
		if err != nil {
			panic(err)
		}
		prod = append(prod, Product{ID: id, Name: name, Quantity: quantity, Price: price})
	}
	return prod
}

func UpdateID(item Product) (int64, error) {
	_, err := GetDataID(item.ID)
	if err != nil {
		return -1, err
	}
	res, err := db.Exec("UPDATE inventory SET name = $2, quantity =$3, price = $4 WHERE id=$1",
		item.ID, item.Name, item.Quantity, item.Price)
	if err != nil {
		return -1, err
	}
	rowcount, err := res.RowsAffected()
	if err != nil {
		return -1, err
	}
	return rowcount, nil
}
func DeleteID(IDNAME string) (int64, error) {
	res, err := db.Exec("DELETE FROM inventory WHERE ID = $1", IDNAME)
	if err != nil {
		return -1, err
	}
	rowcount, err := res.RowsAffected()
	if err != nil {
		return -1, err
	}
	return rowcount, nil
}
