package main

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"github.com/graphql-go/graphql"
	"github.com/lib/pq"
	_ "github.com/lib/pq"
)

type Product struct {
	ID          string  `json:"item_id"`
	Naming      string  `json:"name"`
	Weight      float64 `json:"weight"`
	Description string  `json:"description"`
}
type Cart struct {
	Prods []Product `json:"product"`
}

var (
	cart     Cart
	PortAddr = ":8080"
	db       *sql.DB
)

const (
	database_URL = "host=postgres port=5432 user=postgres password=1234 dbname=productdb sslmode=disable"
)

var ProductType = graphql.NewObject(graphql.ObjectConfig{
	Name: "Product",
	Fields: graphql.Fields{
		"item_id": &graphql.Field{
			Type: graphql.String,
		},
		"name": &graphql.Field{
			Type: graphql.String,
		},
		"weight": &graphql.Field{
			Type: graphql.Float,
		},
		"description": &graphql.Field{
			Type: graphql.String,
		},
	},
})
var rootQuery = graphql.NewObject(graphql.ObjectConfig{
	Name: "RootQuery",
	Fields: graphql.Fields{
		"product": &graphql.Field{
			Type: graphql.NewList(ProductType),
			Resolve: func(p graphql.ResolveParams) (interface{}, error) {
				return cart.Prods, nil
			},
		},
	},
})

func GraphQl() []byte {
	schema, err := graphql.NewSchema(graphql.SchemaConfig{
		Query: rootQuery,
	})
	if err != nil {
		fmt.Println("Error", err)
		return nil
	}
	query := `
		query {
			product {
				item_id
				name
				weight
				description
			}
		}
	`
	params := graphql.Params{
		Schema:        schema,
		RequestString: query,
	}
	r := graphql.Do(params)
	if len(r.Errors) > 0 {
		fmt.Printf("Failed to execute query: %v", r.Errors)
		return nil
	}
	fmt.Println("r-data = ", r.Data)
	jsdata, err := json.Marshal(r.Data)
	fmt.Println("js-data = ", string(jsdata))
	jsonData, err := json.MarshalIndent(r, "", "  ")
	if err != nil {
		fmt.Println("Failed to marshal JSON:", err)
		return nil
	}
	fmt.Println("json-data = ", string(jsonData))
	return jsdata
}

func main() {
	db = ConnectDd()

	defer db.Close()
	fmt.Println("Подключение к PostgreSQL успешно!")
	tables, err := getTables()
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Список таблиц:")
	for _, tableName := range tables {
		fmt.Println(tableName)
	}
	router := mux.NewRouter()
	router.HandleFunc("/products", GetProducts).Methods("GET")           //Получить информацию о всех продуктах
	router.HandleFunc("/products/{id}", GetProduct).Methods("GET")       //Получить информацию о продукте с номером ID
	router.HandleFunc("/products", CreateProduct).Methods("POST")        //Добавить продукт
	router.HandleFunc("/products/{id}", UpdateProduct).Methods("PUT")    //Изменить продукт по ID
	router.HandleFunc("/products/{id}", DeleteProduct).Methods("DELETE") //Удалить продукт ID
	router.HandleFunc("/products/{id}", AddProd).Methods("POST")
	router.HandleFunc("/cart", GETCart).Methods("GET")
	router.HandleFunc("/cart", POSTCart).Methods("POST")

	fmt.Println("Сервер слушате порт " + PortAddr)
	log.Fatal(http.ListenAndServe(PortAddr, router))
}
func getTables() ([]string, error) {
	// Добавляем дополнительную задержку
	time.Sleep(30 * time.Second)

	rows, err := db.Query("SELECT table_name FROM information_schema.tables WHERE table_schema = 'public'")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tables []string
	for rows.Next() {
		var tableName string
		err := rows.Scan(&tableName)
		if err != nil {
			return nil, err
		}
		tables = append(tables, tableName)
	}

	err = rows.Err()
	if err != nil {
		return nil, err
	}

	return tables, nil
}

func Insert(item Product) (int64, error) {
	res, err := db.Exec("INSERT INTO products (id, naming, weight, description) VALUES ($1,$2,$3,$4)", item.ID, item.Naming, item.Weight, item.Description)
	if err != nil {
		return -1, err
	}
	rowcount, err := res.RowsAffected()
	if err != nil {
		return rowcount, err
	}
	return rowcount, nil
}
func UpdateID(item Product) (int64, error) {
	_, err := GetDataID(item.ID)
	if err != nil {
		return -1, err
	}
	res, err := db.Exec("UPDATE products SET naming = $2, weight =$3, description = $4 WHERE id=$1",
		item.ID, item.Naming, item.Weight, item.Description)
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
	res, err := db.Exec("DELETE FROM products WHERE ID = $1", IDNAME)
	if err != nil {
		return -1, err
	}
	rowcount, err := res.RowsAffected()
	if err != nil {
		return -1, err
	}
	return rowcount, nil
}
func GetDataID(IDNAME string) (Product, error) { //Обработать ошибку после работы функции
	rows := db.QueryRow("SELECT * FROM products WHERE id=$1", IDNAME)
	var prod Product
	// Обработка результатов запроса
	err := rows.Scan(&prod.ID, &prod.Naming, &prod.Weight, &prod.Description)
	if err != nil {
		return Product{}, err
	}
	return prod, nil
}
func GetData() []Product {
	rows, err := db.Query("SELECT * FROM products")
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
func ConnectDd() *sql.DB {
	db, err := sql.Open("postgres", database_URL)
	if err != nil {
		fmt.Println(err)
		fmt.Println("КАКИ")

		panic(err)
	}
	check := db.Ping()
	if check != nil {

		fmt.Println(check)
	}
	fmt.Println("СИСИ")
	return db
}
func GetProducts(w http.ResponseWriter, r *http.Request) {
	fmt.Println("КАКА2")
	prods := GetData()
	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(prods)
}
func GetProduct(w http.ResponseWriter, r *http.Request) {
	prod, err := GetDataID(mux.Vars(r)["id"])
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
	}
	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(prod)
}
func CreateProduct(w http.ResponseWriter, r *http.Request) {
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
func UpdateProduct(w http.ResponseWriter, r *http.Request) {
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

func POSTCart(w http.ResponseWriter, r *http.Request) {
	jsonData := GraphQl()
	fmt.Println(string(jsonData))
	resp, err := http.Post("http://localhost:8081/orders", "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()
	w.WriteHeader(http.StatusNoContent)
}
func GETCart(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(cart.Prods)
}
func AddProd(w http.ResponseWriter, r *http.Request) {
	prod, err := GetDataID(mux.Vars(r)["id"])
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
	}
	cart.Prods = append(cart.Prods, prod)
	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(cart.Prods)
}
func DeleteProduct(w http.ResponseWriter, r *http.Request) {
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
