package distancehelper

import (
	"context"
	"errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"googlemaps.github.io/maps"
	"os"
	"request"
	"testing"
	"time"
)

type GMapClientMock struct {
	mock.Mock
	GMapClient
}

func (m *GMapClientMock) DistanceMatrix(ctx context.Context, r *maps.DistanceMatrixRequest) (*maps.DistanceMatrixResponse, error) {
	args := m.Called(ctx, r)
	return args.Get(0).(*maps.DistanceMatrixResponse), args.Error(1)
}

type GMapMock struct {
	mock.Mock
	GMap
}

func (m *GMapMock) GetClient(apiKey string) (GMapClient, error) {
	args := m.Called(apiKey)
	return args.Get(0).(GMapClient), args.Error(1)
}

var req = &request.PlaceOrderRequest{Origin: []string{"22.2802", "114.184919"}, Destination: []string{"25.052192", "121.522333"}}
var gh = &GMapHelper{}

func TestDistanceWithNoKeyAndEmptyRequest(t *testing.T) {
	os.Remove(apiKeyName)
	d, err := gh.GetDistanceMeters(&request.PlaceOrderRequest{}, &GMapMock{})
	if d != 0 || err != nil {
		t.Errorf("Incorrect distance: got %d, expected 0; err: %v", d, err)
	}
}

func TestDistanceWithNoKeyAndNonEmptyRequest(t *testing.T) {
	os.Remove(apiKeyName)
	d, err := gh.GetDistanceMeters(req, &GMapMock{})
	if d != 0 || err != nil {
		t.Errorf("Incorrect distance: got %d, expected 0; err: %v", d, err)
	}
}

func TestEmptyAPIKey(t *testing.T) {
	os.Setenv(apiKeyName, "")

	defer func() {
		if r := recover(); r == nil {
			t.Errorf("The code did not panic")
		}
	}()

	// The following is the code under test
	gh.GetDistanceMeters(req, &GMapMock{})
}

func TestHappyFlow(t *testing.T) {
	os.Setenv(apiKeyName, "A")

	d, _ := gh.GetDistanceMeters(req, mockInterfaces(getNormalResponse(), nil))

	assert.Equal(t, 1049, d)
}

func TestGMapAPIError(t *testing.T) {
	os.Setenv(apiKeyName, "A")

	d, err := gh.GetDistanceMeters(req, mockInterfaces(getNormalResponse(), errors.New("")))

	assert.NotNil(t, err)
	assert.Equal(t, -1, d)
}

func TestGMapReturnNotOK(t *testing.T) {
	os.Setenv(apiKeyName, "A")

	d, err := gh.GetDistanceMeters(req, mockInterfaces(getErrorResponse(), nil))

	assert.Nil(t, err)
	assert.Equal(t, -1, d)
}

func mockInterfaces(expected *maps.DistanceMatrixResponse, err error) GMap {
	gmapClient := GMapClientMock{}
	gmapClient.On("DistanceMatrix", mock.Anything, mock.Anything).Return(expected, err)

	gmap := GMapMock{}
	gmap.On("GetClient", mock.Anything).Return(&gmapClient, nil)

	return &gmap
}

func getNormalResponse() *maps.DistanceMatrixResponse {
	e := getNormalMatrix()
	return getResponse(&e)
}

func getErrorResponse() *maps.DistanceMatrixResponse {
	e := getErrorMatrix()
	return getResponse(&e)
}

func getResponse(element *maps.DistanceMatrixElement) *maps.DistanceMatrixResponse {
	arr := []*maps.DistanceMatrixElement{element}
	r := maps.DistanceMatrixElementsRow{Elements: arr}
	return &maps.DistanceMatrixResponse{
		OriginAddresses:      []string{req.Origin[0]},
		DestinationAddresses: []string{req.Destination[0]},
		Rows:                 []maps.DistanceMatrixElementsRow{r},
	}
}

func getNormalMatrix() maps.DistanceMatrixElement {
	e := maps.DistanceMatrixElement{
		Status:   "OK",
		Duration: time.Duration(416 * time.Second),
		Distance: maps.Distance{Meters: 1049},
	}
	return e
}

func getErrorMatrix() maps.DistanceMatrixElement {
	e := maps.DistanceMatrixElement{
		Status: "ZERO_RESULTS",
	}
	return e
}
