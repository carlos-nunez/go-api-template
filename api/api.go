package api

import (
	"context"
	"encoding/json"
	"io"
	"io/ioutil"
	"net/http"

	"go.mongodb.org/mongo-driver/mongo"
)

type API struct {
	mdb mongo.Database
	ctx context.Context
}

func NewAPI() *API {
	return &API{}
}

func (a *API) Initialize(db mongo.Database, context context.Context) {
	a.mdb = db
	a.ctx = context
}

func (a *API) marshallBody(b interface{}, w http.ResponseWriter, r *http.Request) error {
	body, err := ioutil.ReadAll(io.LimitReader(r.Body, 1048576))

	if err != nil {
		return err
	}
	if err := r.Body.Close(); err != nil {
		return err
	}
	if err := json.Unmarshal(body, &b); err != nil {
		return err
	}

	return err
}
