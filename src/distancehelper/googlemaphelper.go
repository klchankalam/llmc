package distancehelper

import (
	"context"
	"fmt"
	log "github.com/sirupsen/logrus"
	"googlemaps.github.io/maps"
	"os"
	"request"
	"strings"
)

const (
	distNoKey  = 0
	apiKeyName = "GOOGLE_MAP_API_KEY"
)

type GMap interface {
	GetClient(apiKey string) (GMapClient, error)
}
type GMapReal struct{ GMap }

func (m *GMapReal) GetClient(apiKey string) (GMapClient, error) {
	return maps.NewClient(maps.WithAPIKey(apiKey))
}

type GMapClient interface {
	DistanceMatrix(ctx context.Context, r *maps.DistanceMatrixRequest) (*maps.DistanceMatrixResponse, error)
}
type GMapClientReal struct{ GMapClient }

func (m *GMapClientReal) DistanceMatrix(ctx context.Context, r *maps.DistanceMatrixRequest) (*maps.DistanceMatrixResponse, error) {
	return m.DistanceMatrix(ctx, r)
}

func init() {
	if _, present := os.LookupEnv(apiKeyName); !present {
		log.Warn(fmt.Sprintf("Google API key is not set, distance will always be %d.", distNoKey))
	}
}

func GetDistanceMeters(co *request.PlaceOrderRequest, gm GMap) (int, error) {
	key, present := os.LookupEnv(apiKeyName)
	if !present {
		return 0, nil
	}

	// create client
	c, err := gm.GetClient(key)
	if err != nil {
		panic(fmt.Sprintf("fatal error: %s", err))
	}

	// get distance
	r := &maps.DistanceMatrixRequest{Origins: []string{strings.Join(co.Origin, ",")},
		Destinations: []string{strings.Join(co.Destination, ",")}}
	dist, err := c.DistanceMatrix(context.Background(), r)
	if err != nil {
		log.Errorf("Google map API problem: %v", err)
		return -1, err
	}
	if dist.Rows[0].Elements[0].Status != "OK" {
		return -1, nil
	}

	return dist.Rows[0].Elements[0].Distance.Meters, nil
}
