package main

import (
	"encoding/json"
	"fmt"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mysql"
	_ "github.com/jinzhu/gorm/dialects/postgres"
	"github.com/julienschmidt/httprouter"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"
)

var DB *gorm.DB

func main() {
	router := httprouter.New()
	router.POST("/orders", newOrderHandler)
	router.GET("/orders", listOrderHandler)
	router.GET("/", dummyHandler)
	time.Sleep(5 * time.Second)
	log.Println("init db")
	initDb()
	DB.AutoMigrate(&Order{})
	log.Fatal(http.ListenAndServe(":8080", router))
}

func initDb() {
	//db, err := gorm.Open("mysql", "user:password@tcp(db:3306)/db?charset=utf8mb4&parseTime=True")
	db, err := gorm.Open("postgres", "host=db port=5432 user=postgres dbname=postgres password=password sslmode=disable")
	DB = db
	if err != nil {
		panic(fmt.Sprintf("failed to connect database: %s", err.Error()))
	}
	// TODO defer db.Close()
}

func dummyHandler(writer http.ResponseWriter, request *http.Request, _ httprouter.Params) {
	fmt.Fprintf(writer, "Welcome to my website!")
}

type OrderRequest struct {
	Origin      []string
	Destination []string
}

type Order struct {
	ID          uint64    `gorm:"primary_key" json:"id"`
	Distance    uint64    `gorm:"not null" json:"distance"`
	Status      string    `gorm:"type:varchar(10);not null" json:"status"`
	OriginsLat  string    `json:"-"`
	OriginsLong string    `json:"-"`
	DestLat     string    `json:"-"`
	DestLong    string    `json:"-"`
	CreatedAt   time.Time `json:"-"`
	UpdatedAt   time.Time `json:"-"`
}

type ErrorMessage struct {
	error string
}

func listOrderHandler(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	// get query params
	limit, errLimit := strconv.Atoi(getParamOrDefault(r, "limit", "-1"))
	if errLimit != nil || limit < -1 {
		http.Error(w, getErrorJson(fmt.Sprintf("Invalid page %v", errLimit)), http.StatusBadRequest)
		return
	}
	page, errPage := strconv.Atoi(getParamOrDefault(r, "page", "1"))
	if errPage != nil || page < 1 {
		http.Error(w, getErrorJson(fmt.Sprintf("Invalid page %v", errPage)), http.StatusBadRequest)
		return
	}

	// query
	var orders []Order
	DB.Limit(limit).Offset((page - 1) * limit).Find(&orders)

	// return result to user
	if err := json.NewEncoder(w).Encode(orders); err != nil {
		http.Error(w, getErrorJson(fmt.Sprintf("Cannot marshal JSON body: %v", err)), http.StatusBadRequest)
		return
	}
}

func getParamOrDefault(r *http.Request, param string, def string) string {
	p := r.URL.Query().Get(param)
	if p == "" {
		return def
	}
	return p
}

func newOrderHandler(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	if !checkContentType(r, w, "application/json") {
		return
	}

	var err error

	// get body and check JSON
	var orderRequest OrderRequest
	err = json.NewDecoder(r.Body).Decode(&orderRequest)
	if err != nil || len(orderRequest.Origin) != 2 || len(orderRequest.Destination) != 2 {
		http.Error(w, getErrorJson(fmt.Sprintf("Cannot parse JSON body: %v", err)), http.StatusBadRequest)
		return
	}

	if !isLatitude(orderRequest.Origin[0]) || !isLongitude(orderRequest.Origin[1]) ||
		!isLatitude(orderRequest.Destination[0]) || !isLongitude(orderRequest.Destination[1]) {
		http.Error(w,
			getErrorJson(fmt.Sprintf("Incorrect input - must be valid latitudes and longitudes: %v", orderRequest)),
			http.StatusBadRequest)
		return
	}

	// TODO Get distance
	var dist uint64 = 10

	// save orderRequest in db
	res := &Order{Distance: dist, Status: "UNASSIGNED",
		OriginsLat: orderRequest.Origin[0], OriginsLong: orderRequest.Origin[1],
		DestLat: orderRequest.Destination[0], DestLong: orderRequest.Destination[1]}
	DB.Create(res)

	// return result to user
	if err = json.NewEncoder(w).Encode(res); err != nil {
		http.Error(w, getErrorJson(fmt.Sprintf("Cannot marshal JSON body: %v", err)), http.StatusBadRequest)
		return
	}

}

func isLatitude(s string) bool {
	return isNum(s, -90, 90)
}

func isLongitude(s string) bool {
	return isNum(s, -180, 180)
}

func isNum(s string, min float64, max float64) bool {
	n, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return false
	}
	return n > min && n < max
}

func checkContentType(r *http.Request, w http.ResponseWriter, ct string) bool {
	contentType := r.Header.Get("Content-Type")
	if !strings.HasPrefix(contentType, ct) {
		http.Error(w, getErrorJson(http.StatusText(http.StatusUnsupportedMediaType)), http.StatusUnsupportedMediaType)
		return false
	}
	return true
}

func getErrorJson(errMsg string) string {
	s, err := json.Marshal(ErrorMessage{errMsg})
	if err != nil {
		panic(fmt.Sprintf("Cannot marshal json: %s", errMsg))
	}
	return string(s)
}
