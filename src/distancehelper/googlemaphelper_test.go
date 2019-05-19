package distancehelper

import (
	"context"
	"github.com/stretchr/testify/mock"
	"googlemaps.github.io/maps"
	"os"
	"requesthandler"
	"testing"
	"time"
)

type mockNewClient struct {
	mock.Mock
}

func (m *mockNewClient) DistanceMatrix(ctx context.Context, r *maps.DistanceMatrixRequest) (*maps.DistanceMatrixResponse, error) {
	args := m.Called(ctx, r)
	return args.Get(0).(*maps.DistanceMatrixResponse), args.Error(1)
}

var req = &requesthandler.PlaceOrderRequest{Origin: []string{"22.2802", "114.184919"}, Destination: []string{"25.052192", "121.522333"}}

func TestDistanceWithNoKeyAndEmptyRequest(t *testing.T) {
	os.Remove(apiKeyName)
	d, err := GetDistanceMeters(&requesthandler.PlaceOrderRequest{})
	if d != 0 || err != nil {
		t.Errorf("Incorrect distance: got %d, expected 0; err: %v", d, err)
	}
}

func TestDistanceWithNoKeyAndNonEmptyRequest(t *testing.T) {
	os.Remove(apiKeyName)
	d, err := GetDistanceMeters(req)
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
	GetDistanceMeters(req)
}

func TestHappyFlow(t *testing.T) {
	os.Setenv(apiKeyName, "A")

	testObj := new(mockNewClient)
	testObj.On("DistanceMatrix", mock.Anything).Return(getNormalResponse(), nil)

	GetDistanceMeters(req)

	testObj.AssertExpectations(t)
}

func getNormalResponse() interface{} {
	e := maps.DistanceMatrixElement{
		Status:   "OK",
		Duration: time.Duration(416 * time.Second),
		Distance: maps.Distance{Meters: 1049},
	}
	arr := []*maps.DistanceMatrixElement{&e}
	r := maps.DistanceMatrixElementsRow{Elements: arr}
	return &maps.DistanceMatrixResponse{
		OriginAddresses:      []string{req.Origin[0]},
		DestinationAddresses: []string{req.Destination[0]},
		Rows:                 []maps.DistanceMatrixElementsRow{r},
	}
}
