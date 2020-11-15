package contactcache

import (
	"encoding/json"
	"net/http"
)

type errorResp struct {
	ErrorType string `json:"error"`
	Message   string `json:"message"`
}

func httpJSONError(w http.ResponseWriter, msg string, code int) {
	err := &errorResp{
		ErrorType: http.StatusText(code),
		Message:   msg,
	}

	json.NewEncoder(w).Encode(err)
}
