package requesthandler

import (
	"fmt"
	"net/http"
	"responseutil"
	"strconv"
	"strings"
)

func getPageAndLimit(req *http.Request) (int, int, []string) {
	limitMin, pageMin, pageDefault := 1, 1, 1
	limitDefault := -1
	var es []string

	limit, err := getNumberFromRequestWithLowerBound(req, "limit", limitDefault, limitMin)
	if err != "" {
		es = append(es, err)
	}
	page, err := getNumberFromRequestWithLowerBound(req, "page", pageDefault, pageMin)
	if err != "" {
		es = append(es, err)
	}

	return page, limit, es
}

// return param in number, in case of err, default value and err are returned
func getNumberFromRequestWithLowerBound(req *http.Request, param string, def int, min int) (int, string) {
	num, err := strconv.Atoi(getParamOrDefault(req, param, def))
	if err != nil {
		return def, fmt.Sprintf("Invalid %s %v", param, err)
	}
	if num != def && num < min {
		return def, fmt.Sprintf("Invalid %s %d", param, num)
	}
	return num, ""
}

func getParamOrDefault(r *http.Request, param string, def int) string {
	p := r.URL.Query().Get(param)
	if p == "" {
		p = strconv.Itoa(def)
	}
	return p
}

func coordinatesValid(request PlaceOrderRequest) bool {
	return !isLatitude(request.Origin[0]) || !isLongitude(request.Origin[1]) ||
		!isLatitude(request.Destination[0]) || !isLongitude(request.Destination[1])
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
		responseutil.WriteJSONErrorResponse(w, http.StatusText(http.StatusUnsupportedMediaType), http.StatusUnsupportedMediaType)
		return false
	}
	return true
}
