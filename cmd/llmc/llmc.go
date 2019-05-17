package main

import (
	"encoding/json"
	"fmt"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mysql"
	_ "github.com/jinzhu/gorm/dialects/postgres"
	"github.com/julienschmidt/httprouter"
	log "github.com/sirupsen/logrus"
	"net/http"
	"strconv"
	"strings"
	"time"
)

var DB *gorm.DB

const (
	StatusUnassigned = "UNASSIGNED"
	StatusTaken      = "TAKEN"
)

func main() {
	// setup routes
	router := httprouter.New()
	router.POST("/orders", newOrderHandler)
	router.PATCH("/orders/:id", takeOrderHandler)
	router.GET("/orders", listOrderHandler)

	// setup db
	// TODO use wait?
	time.Sleep(2 * time.Second)
	log.Println("initializing DB...")
	initDb()
	defer DB.Close()
	DB.AutoMigrate(&Order{})
	log.Println("DB initialized")

	// start server
	log.Fatal(http.ListenAndServe(":8080", router))
	log.Println("Server started")
}

func initDb() {
	//db, err := gorm.Open("mysql", "user:password@tcp(db:3306)/db?charset=utf8mb4&parseTime=True")
	db, err := gorm.Open("postgres", "host=db port=5432 user=postgres dbname=postgres password=password sslmode=disable")
	DB = db
	if err != nil {
		panic(fmt.Sprintf("failed to connect database: %s", err.Error()))
	}
}

type PlaceOrderRequest struct {
	Origin      []string
	Destination []string
}

type TakeOrder struct {
	Status string `json:"Status"`
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
	Error string `json:"error"`
}

func listOrderHandler(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	// get query params
	limit, errLimit := strconv.Atoi(getParamOrDefault(r, "limit", "-1"))
	if errLimit != nil || limit < -1 {
		writeJSONErrorResponse(w, fmt.Sprintf("Invalid limit %v", errLimit), http.StatusBadRequest)
		return
	}
	page, errPage := strconv.Atoi(getParamOrDefault(r, "page", "1"))
	if errPage != nil || page < 1 {
		writeJSONErrorResponse(w, fmt.Sprintf("Invalid page %v", errPage), http.StatusBadRequest)
		return
	}

	// query
	var orders []Order
	DB.Limit(limit).Offset((page - 1) * limit).Find(&orders)

	// return result to user
	writeJSONToResponse(&orders, w)
}

func getParamOrDefault(r *http.Request, param string, def string) string {
	p := r.URL.Query().Get(param)
	if p == "" {
		return def
	}
	return p
}

func takeOrderHandler(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	// check input
	ids := ps.ByName("id")
	id, err := strconv.Atoi(ids)
	if err != nil || id < 1 {
		writeJSONErrorResponse(w, fmt.Sprintf("Invalid Id: %s", ids), http.StatusBadRequest)
		return
	}

	// get entity
	var order Order
	DB.Where("status = ?", StatusUnassigned).First(&order, id)
	if order.ID == 0 {
		writeJSONErrorResponse(w, fmt.Sprintf("Order id %d with status %s not found", id, StatusUnassigned), http.StatusNotFound)
		return
	}

	// get body and check JSON
	var jsonReq TakeOrder
	err = json.NewDecoder(r.Body).Decode(&jsonReq)
	if err != nil {
		writeJSONErrorResponse(w, fmt.Sprintf("Cannot parse JSON body: %v", err), http.StatusBadRequest)
		return
	}
	// only accept taken as status
	if jsonReq.Status != StatusTaken {
		writeJSONErrorResponse(w, "Invalid request status", http.StatusBadRequest)
		return
	}

	// to avoid multiple updates, we add the where check
	updateResult := DB.Model(&order).Where("Status = ?", StatusUnassigned).Update("Status", StatusTaken)
	if updateResult.RowsAffected < 1 {
		if updateResult.Error != nil {
			writeJSONErrorResponse(w, fmt.Sprintf("Update error: %v", updateResult.Error), http.StatusBadRequest)
		} else {
			writeJSONErrorResponse(w, "Not updated - perhaps updated moment ago?", http.StatusBadRequest)
		}
	} else {
		writeJSONToResponse(&TakeOrder{"SUCCESS"}, w)
	}
}

func newOrderHandler(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	if !checkContentType(r, w, "application/json") {
		return
	}

	// get body and check JSON
	var orderRequest PlaceOrderRequest
	err := json.NewDecoder(r.Body).Decode(&orderRequest)
	if err != nil || len(orderRequest.Origin) != 2 || len(orderRequest.Destination) != 2 {
		writeJSONErrorResponse(w, fmt.Sprintf("Cannot parse JSON body: %v", err), http.StatusBadRequest)
		return
	}

	if !isLatitude(orderRequest.Origin[0]) || !isLongitude(orderRequest.Origin[1]) ||
		!isLatitude(orderRequest.Destination[0]) || !isLongitude(orderRequest.Destination[1]) {
		writeJSONErrorResponse(w, fmt.Sprintf("Incorrect input - must be valid latitudes and longitudes: %v", orderRequest), http.StatusBadRequest)
		return
	}

	// TODO Get distance
	var dist uint64 = 10

	// save orderRequest in db
	res := &Order{Distance: dist, Status: "UNASSIGNED",
		OriginsLat: orderRequest.Origin[0], OriginsLong: orderRequest.Origin[1],
		DestLat: orderRequest.Destination[0], DestLong: orderRequest.Destination[1]}
	createResult := DB.Create(res)
	if createResult.Error != nil || res.ID == 0 {
		writeJSONErrorResponse(w, fmt.Sprintf("Create error: %v", createResult.Error), http.StatusBadRequest)
		return
	}

	// return result to user
	writeJSONToResponse(&res, w)
}

func isLatitude(s string) bool {
	return isNumWithRange(s, -90, 90)
}

func isLongitude(s string) bool {
	return isNumWithRange(s, -180, 180)
}

func isNumWithRange(s string, min float64, max float64) bool {
	n, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return false
	}
	return n > min && n < max
}

func checkContentType(r *http.Request, w http.ResponseWriter, ct string) bool {
	contentType := r.Header.Get("Content-Type")
	if !strings.HasPrefix(contentType, ct) {
		writeJSONErrorResponse(w, http.StatusText(http.StatusUnsupportedMediaType), http.StatusUnsupportedMediaType)
		return false
	}
	return true
}

func getErrorJsonString(errMsg string) string {
	s, err := json.Marshal(&ErrorMessage{errMsg})
	if err != nil {
		panic(fmt.Sprintf("Cannot marshal json: %s", errMsg))
	}
	return string(s)
}

func writeJSONToResponse(v interface{}, w http.ResponseWriter) {
	if err := json.NewEncoder(w).Encode(v); err != nil {
		writeJSONErrorResponse(w, fmt.Sprintf("Cannot marshal JSON body: %v", err), http.StatusInternalServerError)
		return
	}
}

func writeJSONErrorResponse(w http.ResponseWriter, error string, code int) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.WriteHeader(code)
	_, _ = fmt.Fprintln(w, getErrorJsonString(error))
}
