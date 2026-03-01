package v1

import "net/http"

type DummyController interface {
	GetDummy(w http.ResponseWriter, r *http.Request)
}

type dummyController struct{}

func NewDummyController() *dummyController {
	return &dummyController{}
}

func (d *dummyController) GetDummy(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(200)
}
