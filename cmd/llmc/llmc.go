package main

import (
	//"github.com/jinzhu/gorm"
	"encoding/json"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"log"
	"net/http"
	"strings"
)

func main() {
	http.HandleFunc("/orders", newOrderHandler)
	http.HandleFunc("/", dummyHandler)
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func dummyHandler(writer http.ResponseWriter, request *http.Request) {
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

func newOrderHandler(w http.ResponseWriter, r *http.Request) {
	if !checkMethod(r, w, http.MethodPost) {
		return
	}
	if !checkContentType(r, w, "application/json") {
		return
	}

	// get body and check JSON
	var order OrderRequest
	jsonErr := json.NewDecoder(r.Body).Decode(&order)
	if jsonErr != nil || len(order.origin) != 2 || len(order.destination) != 2 {
		writeHTTPErrorHeader(w, http.StatusBadRequest, fmt.Sprintf("Cannot parse JSON body: %s", jsonErr.Error()))
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
		writeHTTPErrorHeader(w, http.StatusMethodNotAllowed, http.StatusText(http.StatusMethodNotAllowed))
		return false
	}
	return true
}

func checkContentType(r *http.Request, w http.ResponseWriter, ct string) bool {
	contentType := r.Header.Get("Content-Type")
	if !strings.HasPrefix(contentType, ct) {
		writeHTTPErrorHeader(w, http.StatusUnsupportedMediaType, http.StatusText(http.StatusUnsupportedMediaType))
		return false
	}
	return true
}

func writeHTTPErrorHeader(w http.ResponseWriter, e int, s string) {
	w.WriteHeader(e)
	w.Write([]byte(s))
}
