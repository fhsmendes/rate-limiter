package handler

import (
	"net/http"
)

type TestHandler struct {
}

func NewTestHandler() *TestHandler {
	return &TestHandler{}
}

func (t *TestHandler) GetTest(w http.ResponseWriter, r *http.Request) {

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

}
