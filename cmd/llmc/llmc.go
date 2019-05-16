package main

import (
	//"github.com/jinzhu/gorm"
	"encoding/json"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"github.com/julienschmidt/httprouter"
	"log"
	"net/http"
	"strings"
)

func main() {
	router := httprouter.New()
	router.POST("/orders", newOrderHandler)
	router.GET("/", dummyHandler)
	log.Fatal(http.ListenAndServe(":8080", router))
}

func dummyHandler(writer http.ResponseWriter, request *http.Request, _ httprouter.Params) {
	fmt.Fprintf(writer, "Welcome to my website!")
}

type OrderRequest struct {
	origin      []string
	destination []string
}

type OrderResponse struct {
	id       uint64 `gorm:"primary_key"`
	distance float64
	status   string `gorm:"size:10"`
}

func newOrderHandler(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	if !checkContentType(r, w, "application/json") {
		return
	}

	// get body and check JSON
	var order OrderRequest
	jsonErr := json.NewDecoder(r.Body).Decode(&order)
	if jsonErr != nil || len(order.origin) != 2 || len(order.destination) != 2 {
		http.Error(w, fmt.Sprintf("Cannot parse JSON body: %s", jsonErr), http.StatusBadRequest)
		return
	}

	// TODO Get distance
	dist := 10.0

	// TODO save order in db
	var id uint64 = 1

	// return result to user
	res := OrderResponse{id, dist, "UNASSIGNED"}
	json.NewEncoder(w).Encode(res)
}

func checkMethod(r *http.Request, w http.ResponseWriter, method string) bool {
	if strings.ToUpper(r.Method) != method {
		http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
		return false
	}
	return true
}

func checkContentType(r *http.Request, w http.ResponseWriter, ct string) bool {
	contentType := r.Header.Get("Content-Type")
	if !strings.HasPrefix(contentType, ct) {
		http.Error(w, http.StatusText(http.StatusUnsupportedMediaType), http.StatusUnsupportedMediaType)
		return false
	}
	return true
}
