package requesthandler

import (
	db "dao"
	"distancehelper"
	"encoding/json"
	"entity"
	"fmt"
	"github.com/jinzhu/gorm"
	"github.com/julienschmidt/httprouter"
	"net/http"
	"request"
	"responseutil"
	"strconv"
	"strings"
)

const (
	StatusUnassigned = "UNASSIGNED"
	StatusTaken      = "TAKEN"
)

type TakeOrder struct {
	Status string `json:"Status"`
}

type Dependencies struct {
	DB  *gorm.DB
	Dao db.DAO
	Map distancehelper.GMap
}

func (dep *Dependencies) HandleListOrder(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	// get query params
	page, limit, err := getPageAndLimit(r)
	if len(err) > 0 {
		responseutil.WriteJSONErrorResponse(w, strings.Join(err, "; "), http.StatusBadRequest)
		return
	}

	// query
	var orders []entity.Order
	dep.Dao.FindWithLimitAndOffset(dep.DB, limit, (page-1)*limit, &orders)

	// return result to user
	responseutil.WriteJSONToResponse(&orders, w)
}

func (dep *Dependencies) HandleTakeOrder(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	// check input
	ids := ps.ByName("id")
	id, err := strconv.Atoi(ids)
	if err != nil || id < 1 {
		responseutil.WriteJSONErrorResponse(w, fmt.Sprintf("Invalid Id: %s", ids), http.StatusBadRequest)
		return
	}

	// get entity
	var order entity.Order
	dep.Dao.FindFirstWithIdAndStatus(dep.DB, StatusUnassigned, id, &order)
	if order.ID == 0 {
		responseutil.WriteJSONErrorResponse(w, fmt.Sprintf("Order id %d with status %s not found", id, StatusUnassigned), http.StatusNotFound)
		return
	}

	// get body and check JSON
	var jsonReq TakeOrder
	err = json.NewDecoder(r.Body).Decode(&jsonReq)
	if err != nil {
		responseutil.WriteJSONErrorResponse(w, fmt.Sprintf("Cannot parse JSON body: %v", err), http.StatusBadRequest)
		return
	}
	// only accept taken as status
	if jsonReq.Status != StatusTaken {
		responseutil.WriteJSONErrorResponse(w, "Invalid request status", http.StatusBadRequest)
		return
	}

	// to avoid multiple updates, we add the where check
	updateResult := dep.Dao.UpdateOrderStatus(dep.DB, &order, StatusTaken, StatusUnassigned)
	if updateResult.RowsAffected < 1 {
		if updateResult.Error != nil {
			responseutil.WriteJSONErrorResponse(w, fmt.Sprintf("Update error: %v", updateResult.Error), http.StatusBadRequest)
		} else {
			responseutil.WriteJSONErrorResponse(w, "Not updated - perhaps updated moment ago?", http.StatusBadRequest)
		}
		return
	} else {
		responseutil.WriteJSONToResponse(&TakeOrder{"SUCCESS"}, w)
	}
}

func (dep *Dependencies) HandleNewOrder(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	if !checkContentType(r, w, "application/json") {
		return
	}

	// get body and check JSON
	var orderRequest request.PlaceOrderRequest
	err := json.NewDecoder(r.Body).Decode(&orderRequest)
	if err != nil || len(orderRequest.Origin) != 2 || len(orderRequest.Destination) != 2 {
		responseutil.WriteJSONErrorResponse(w, fmt.Sprintf("Cannot parse JSON body: %v", err), http.StatusBadRequest)
		return
	}
	// check coordinates
	if !coordinatesValid(&orderRequest) {
		responseutil.WriteJSONErrorResponse(w, fmt.Sprintf("Incorrect input - must be valid latitudes and longitudes: %v", &orderRequest), http.StatusBadRequest)
		return
	}

	// Get distance
	dist, err := distancehelper.GetDistanceMeters(&orderRequest, dep.Map)
	if err != nil {
		responseutil.WriteJSONErrorResponse(w, fmt.Sprintf("Canno find distance: %v", err), http.StatusInternalServerError)
		return
	}
	if dist == -1 {
		responseutil.WriteJSONErrorResponse(w, "Canno find distance, please check your input.", http.StatusBadRequest)
		return
	}

	// save orderRequest in db
	res := &entity.Order{Distance: dist, Status: "UNASSIGNED",
		OriginsLat: orderRequest.Origin[0], OriginsLong: orderRequest.Origin[1],
		DestLat: orderRequest.Destination[0], DestLong: orderRequest.Destination[1]}
	createResult := dep.Dao.CreateOrder(dep.DB, res)
	if createResult.Error != nil || res.ID == 0 {
		responseutil.WriteJSONErrorResponse(w, fmt.Sprintf("Create error: %v", createResult.Error), http.StatusBadRequest)
		return
	}

	// return result to user
	responseutil.WriteJSONToResponse(&res, w)
}
