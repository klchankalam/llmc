package requesthandler

import (
	"dao"
	"distancehelper"
	"encoding/json"
	"entity"
	"errors"
	"fmt"
	"github.com/jinzhu/gorm"
	"github.com/julienschmidt/httprouter"
	"github.com/stretchr/testify/mock"
	"net/http"
	"net/http/httptest"
	"net/url"
	"request"
	"strconv"
	"strings"
	"testing"
)

// new interfaces and structs for mocks

type GormDBMock struct {
	mock.Mock
	dao.DAO
}

func (gdb *GormDBMock) FindWithLimitAndOffset(db *gorm.DB, limit int, offset int, out *[]entity.Order) {
	args := gdb.Called(db, limit, offset, out)
	*out = *args.Get(0).(*[]entity.Order)
}

func (gdb *GormDBMock) FindFirstWithIdAndStatus(db *gorm.DB, status string, id int, out *entity.Order) {
	args := gdb.Called(db, status, id, out)
	if args.Bool(0) {
		*out = *args.Get(1).(*entity.Order)
	}
}

func (gdb *GormDBMock) UpdateOrderStatus(db *gorm.DB, modelToUpdate *entity.Order, newStatus string, oldStatus string) *gorm.DB {
	args := gdb.Called(db, modelToUpdate, newStatus, oldStatus)
	if modelToUpdate.Status == oldStatus {
		modelToUpdate.Status = newStatus
	}
	return args.Get(0).(*gorm.DB)
}

func (gdb *GormDBMock) CreateOrder(db *gorm.DB, modelToCreate *entity.Order) *gorm.DB {
	args := gdb.Called(db, modelToCreate)
	modelToCreate.ID = args.Get(1).(uint64)
	return args.Get(0).(*gorm.DB)
}

type GMapHelperMock struct {
	mock.Mock
	distancehelper.MapHelper
}

func (ghm *GMapHelperMock) GetDistanceMeters(co *request.PlaceOrderRequest, gm distancehelper.GMap) (int, error) {
	args := ghm.Called(co, gm)
	return args.Get(0).(int), args.Error(1)
}

// test list order

func TestListOrder(t *testing.T) {
	// Arrange
	var h http.Request
	h.URL = &url.URL{RawQuery: ""}
	w := httptest.NewRecorder()
	dao := &GormDBMock{}
	dep := &Dependencies{DB: nil, Map: nil, Dao: dao}
	o := entity.Order{
		ID: 10, Status: StatusUnassigned, Distance: 100,
	}
	dao.On("FindWithLimitAndOffset", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(&[]entity.Order{o})

	//Act
	dep.HandleListOrder(w, &h, nil)

	//Assert
	checkNonEmptyResponse(t, w, http.StatusOK)
	var b []entity.Order
	_ = json.NewDecoder(w.Body).Decode(&b)
	if len(b) != 1 {
		t.Errorf("Expected return length 1, got %d", len(b))
	}
	if b[0].ID != o.ID || b[0].Status != o.Status || b[0].Distance != o.Distance {
		t.Errorf("Expect object is %#v, got %#v", o, b[0])
	}
}

func TestListOrderInputError(t *testing.T) {
	var h http.Request
	h.URL = &url.URL{RawQuery: "page=-2"}
	w := httptest.NewRecorder()
	dep := &Dependencies{DB: nil, Map: nil, Dao: nil}

	dep.HandleListOrder(w, &h, nil)

	checkNonEmptyResponse(t, w, http.StatusBadRequest)
}

// New order tests
var normalCoordinates = "{\"origin\": [\"22.2802\", \"114.184919\"], \"destination\": [\"22.280457\", \"114.185672\"]}"
var id = 10
var distance = 73

func TestNewOrderContentTypeError(t *testing.T) {
	var h http.Request
	w := httptest.NewRecorder()
	dep := &Dependencies{DB: nil, Map: nil, Dao: nil}

	dep.HandleNewOrder(w, &h, nil)

	checkNonEmptyResponse(t, w, http.StatusUnsupportedMediaType)
}

func TestNewOrderJSONError(t *testing.T) {
	testNewOrder(t, strings.NewReader("{\"origin\": [\"-18"), nil, nil, http.StatusBadRequest)
}

func TestNewOrderCoordinatesError(t *testing.T) {
	testNewOrder(t, strings.NewReader("{\"origin\": [\"-180.1\", \"1\"], \"destination\": [\"1\", \"1\"]}"), nil, nil, http.StatusBadRequest)
}

func TestNewOrderMapAPIError(t *testing.T) {
	ghm := getMockMapForNewOrder(-1, errors.New(""))
	testNewOrder(t, strings.NewReader(normalCoordinates), ghm, nil, http.StatusInternalServerError)
}

func TestNewOrderMapNotOKError(t *testing.T) {
	ghm := getMockMapForNewOrder(-1, nil)
	testNewOrder(t, strings.NewReader(normalCoordinates), ghm, nil, http.StatusBadRequest)
}

func TestNewOrderDBError(t *testing.T) {
	ghm := getMockMapForNewOrder(distance, nil)
	dao := getMockDaoForNewOrder(id, errors.New(""))

	testNewOrder(t, strings.NewReader(normalCoordinates), ghm, dao, http.StatusInternalServerError)
}

func TestNewOrderCreateIDError(t *testing.T) {
	ghm := getMockMapForNewOrder(distance, nil)
	dao := getMockDaoForNewOrder(0, nil)

	testNewOrder(t, strings.NewReader(normalCoordinates), ghm, dao, http.StatusInternalServerError)
}

func TestNewOrder(t *testing.T) {
	ghm := getMockMapForNewOrder(distance, nil)
	dao := getMockDaoForNewOrder(id, nil)
	expectedOrder := entity.Order{
		ID: uint64(id), Status: StatusUnassigned, Distance: distance,
	}

	w := testNewOrder(t, strings.NewReader(normalCoordinates), ghm, dao, http.StatusOK)

	var order entity.Order
	_ = json.NewDecoder(w.Body).Decode(&order)
	if order.ID != expectedOrder.ID || order.Status != expectedOrder.Status || order.Distance != expectedOrder.Distance {
		t.Errorf("Expect object is %#v, got %#v", expectedOrder, order)
	}
}

func testNewOrder(t *testing.T, body *strings.Reader, m distancehelper.MapHelper, dao dao.DAO, status int) (w *httptest.ResponseRecorder) {
	r, _ := http.NewRequest("POST", "/orders", body)
	r.Header = map[string][]string{
		"Content-Type": {"application/json", "text/utf-8"},
	}
	w = httptest.NewRecorder()
	dep := &Dependencies{Dao: dao, MapHelper: m}

	dep.HandleNewOrder(w, r, nil)

	checkNonEmptyResponse(t, w, status)

	return w
}

func checkNonEmptyResponse(t *testing.T, w *httptest.ResponseRecorder, status int) {
	if w.Code != status {
		t.Errorf("Expect status %d, actual: %d", status, w.Code)
	}
	if !strings.Contains(w.Header().Get("Content-Type"), "application/json") {
		t.Errorf("Expect return content type is %s, actual: %s", "application/json", w.Header().Get("Content-Type"))
	}
	if w.Body.String() == "" {
		t.Errorf("Expect non empty json response.")
	}
}

func getMockMapForNewOrder(distance int, err error) *GMapHelperMock {
	ghm := &GMapHelperMock{}
	ghm.On("GetDistanceMeters", mock.Anything, mock.Anything).Return(distance, err)
	return ghm
}

func getMockDaoForNewOrder(id int, err error) *GormDBMock {
	dao := &GormDBMock{}
	dao.On("CreateOrder", mock.Anything, mock.Anything).Return(&gorm.DB{Error: err}, uint64(id))
	return dao
}

// take order test
var order = &entity.Order{ID: uint64(id), Status: StatusUnassigned, Distance: distance}

func TestTakeOrderIDInvalid(t *testing.T) {
	testTakeOrder(t, "asdf", nil, http.StatusBadRequest, strings.NewReader("{}"))
}

func TestTakeOrderIDNegative(t *testing.T) {
	testTakeOrder(t, strconv.Itoa(-1), nil, http.StatusBadRequest, strings.NewReader("{}"))
}

func TestTakeOrderUnassignedNotFound(t *testing.T) {
	testTakeOrder(t, strconv.Itoa(id), getMockDaoForTakeOrder(nil, nil), http.StatusNotFound, strings.NewReader("{}"))
}

func TestTakeOrderJSONIncorrect(t *testing.T) {
	testTakeOrder(t, strconv.Itoa(id), getMockDaoForTakeOrder(order, nil), http.StatusBadRequest, strings.NewReader("{\"asdf\":"))
}

func TestTakeOrderJSONStatusIncorrect(t *testing.T) {
	testTakeOrder(t, strconv.Itoa(id), getMockDaoForTakeOrder(order, nil), http.StatusBadRequest, strings.NewReader("{\"status\":\"asdf\""))
}

func TestTakeOrderUpdateError(t *testing.T) {
	testTakeOrder(t, strconv.Itoa(id), getMockDaoForTakeOrder(order, &gorm.DB{RowsAffected: 0, Error: errors.New("")}), http.StatusInternalServerError, strings.NewReader("{\"status\":\"TAKEN\"}"))
}

func TestTakeOrderUpdateFailed(t *testing.T) {
	testTakeOrder(t, strconv.Itoa(id), getMockDaoForTakeOrder(order, &gorm.DB{RowsAffected: 0}), http.StatusBadRequest, strings.NewReader("{\"status\":\"TAKEN\"}"))
}

func TestTakeOrderOK(t *testing.T) {
	w := testTakeOrder(t, strconv.Itoa(id), getMockDaoForTakeOrder(order, &gorm.DB{RowsAffected: 1}), http.StatusOK, strings.NewReader("{\"status\":\"TAKEN\"}"))

	if !strings.Contains(w.Body.String(), StatusSuccess) {
		t.Errorf("Expected success result, got %s", w.Body.String())
	}
}

func testTakeOrder(t *testing.T, id string, dao dao.DAO, status int, body *strings.Reader) (w *httptest.ResponseRecorder) {
	r, _ := http.NewRequest("PATCH", fmt.Sprintf("/orders/%s", id), body)
	w = httptest.NewRecorder()
	dep := &Dependencies{Dao: dao}
	params := httprouter.Params{
		httprouter.Param{
			Key: "id", Value: id,
		},
	}

	dep.HandleTakeOrder(w, r, params)

	checkNonEmptyResponse(t, w, status)

	return w
}

func getMockDaoForTakeOrder(order *entity.Order, updateResult *gorm.DB) *GormDBMock {
	dao := &GormDBMock{}
	dao.On("FindFirstWithIdAndStatus", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(order != nil, order)
	dao.On("UpdateOrderStatus", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(updateResult)
	return dao
}
