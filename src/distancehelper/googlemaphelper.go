package distancehelper

import (
	"context"
	"fmt"
	log "github.com/sirupsen/logrus"
	"googlemaps.github.io/maps"
	"os"
	"requesthandler"
	"strings"
)

const (
	distNoKey  = 0
	apiKeyName = "GOOGLE_MAP_API_KEY"
)

type GMapInputer interface {
	WithAPIKey(apiKey string) maps.ClientOption
	NewClient(options ...maps.ClientOption) (*maps.Client, error)
}

type GMapInput struct {
	WithAPIKeyFunc func(apiKey string) maps.ClientOption
	NewClientFunc  func(options ...maps.ClientOption) (*GMapClientInput, error)
}

type GMapClientInputer interface {
	DistanceMatrix(ctx context.Context, r *maps.DistanceMatrixRequest) (*maps.DistanceMatrixResponse, error)
}

type GMapClientInput struct {
	DistanceMatrixFunc func(ctx context.Context, r *maps.DistanceMatrixRequest) (*maps.DistanceMatrixResponse, error)
}

func init() {
	if _, present := os.LookupEnv(apiKeyName); !present {
		log.Warn(fmt.Sprintf("Google API key is not set, distance will always be %d.", distNoKey))
	}
}

func GetDistanceMeters(co *requesthandler.PlaceOrderRequest, mi *GMapInput) (int, error) {
	key, present := os.LookupEnv(apiKeyName)
	if !present {
		return 0, nil
	}

	// create client
	c, err := mi.NewClientFunc(mi.WithAPIKeyFunc(key))
	if err != nil {
		panic(fmt.Sprintf("fatal error: %s", err))
	}

	// get distance
	r := &maps.DistanceMatrixRequest{Origins: []string{strings.Join(co.Origin, ",")},
		Destinations: []string{strings.Join(co.Destination, ",")}}
	dist, err := c.DistanceMatrixFunc(context.Background(), r)
	if err != nil {
		log.Errorf("Google map API problem: %v", err)
		return -1, err
	}
	if dist.Rows[0].Elements[0].Status != "OK" {
		return -1, nil
	}

	return dist.Rows[0].Elements[0].Distance.Meters, nil
}
