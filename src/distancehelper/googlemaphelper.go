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

func init() {
	if _, present := os.LookupEnv(apiKeyName); !present {
		log.Warn(fmt.Sprintf("Google API key is not set, distance will always be %d.", distNoKey))
	}
}

func GetDistanceMeters(co *requesthandler.PlaceOrderRequest) (int, error) {
	key, present := os.LookupEnv(apiKeyName)
	if !present {
		return 0, nil
	}

	// create client
	c, err := maps.NewClient(maps.WithAPIKey(key))
	if err != nil {
		panic(fmt.Sprintf("fatal error: %s", err))
	}

	// get distance
	r := &maps.DistanceMatrixRequest{Origins: []string{strings.Join(co.Origin, ",")},
		Destinations: []string{strings.Join(co.Destination, ",")}}
	dist, err := c.DistanceMatrix(context.Background(), r)
	if err != nil {
		log.Errorf("Google map API problem: %v", err)
	}
	if dist.Rows[0].Elements[0].Status != "OK" {
		return -1, nil
	}

	return dist.Rows[0].Elements[0].Distance.Meters, nil
}
