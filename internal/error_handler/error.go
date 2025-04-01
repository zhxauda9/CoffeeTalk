package error_handler

import (
	"encoding/json"
	"net/http"

	"hot-coffee/models"
)

func Error(w http.ResponseWriter, ErrorText string, code int) {
	Error := models.Error{
		Code:    code,
		Message: ErrorText,
	}
	jsondata, err := json.MarshalIndent(Error, "", "    ")
	if err != nil {
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	w.Write(jsondata)
}
