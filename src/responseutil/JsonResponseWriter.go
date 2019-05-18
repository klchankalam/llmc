package responseutil

import (
	"encoding/json"
	"fmt"
	"net/http"
)

type ErrorMessage struct {
	Error string `json:"error"`
}

func WriteJSONToResponse(v interface{}, w http.ResponseWriter) {
	setResponseHeaderToJson(w, http.StatusOK)
	if err := json.NewEncoder(w).Encode(v); err != nil {
		WriteJSONErrorResponse(w, fmt.Sprintf("Cannot marshal JSON body: %v", err), http.StatusInternalServerError)
		return
	}
}

func WriteJSONErrorResponse(w http.ResponseWriter, error string, code int) {
	setResponseHeaderToJson(w, code)
	_, _ = fmt.Fprintln(w, getErrorJsonString(error))
}

func setResponseHeaderToJson(w http.ResponseWriter, code int) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.WriteHeader(code)
}

func getErrorJsonString(errMsg string) string {
	s, err := json.Marshal(&ErrorMessage{errMsg})
	if err != nil {
		panic(fmt.Sprintf("Cannot marshal json: %s", errMsg))
	}
	return string(s)
}
