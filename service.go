package metricsum

//go:generate mockgen -source=service.go -destination=mocks/RequestCtx.go -package=mocks

import (
	"encoding/json"
	"fmt"
	"math"
	"net/http"
	"time"

	"github.com/valyala/fasthttp"
)

const routerKey = "key"

var (
	strContentType     = []byte("Content-Type")
	strApplicationJSON = []byte("application/json")
)

// RequestCtx contains methods needed by fasthttp.RequestCtx to allow mocks for unit testing
type RequestCtx interface {
	UserValue(key string) interface{}
	PostBody() []byte
	Error(msg string, statusCode int)
	Time() time.Time
}

// use saver to mock the call to saveKeyValue
type saver interface {
	saveKeyValue(interface{}, int, time.Time, deleter)
}
type summer interface {
	getSum(interface{}) (int, error)
}

// Value allows for easy unmarshalling of POST /metric/{key} body data
type Value struct {
	Value float64 `json:"value"`
}

// Metric allows the saving of a metric value with the associated key e.g. POST /metric/{key} { "value" = 4 }
func (km *KeysMap) Metric(ctx *fasthttp.RequestCtx) {
	metric(ctx, km, km)
}

// metric is the unit testable logic for the Metric function
func metric(ctx RequestCtx, save saver, delete deleter) {

	key := ctx.UserValue(routerKey)

	value := &Value{}
	if err := json.Unmarshal(ctx.PostBody(), value); err != nil {
		ctx.Error(fmt.Sprintf("invalid body JSON: %v", err), http.StatusBadRequest)
		// documented behavior in fasthttp, must reset response headers after calling ctx.Error
		setRespContentType(ctx)
		return
	}

	val := int(math.Round(value.Value))

	if val == 0 {
		ctx.Error("could not interpret value - ensure JSON is correct and that value does not round to zero to be interepreted", http.StatusBadRequest)
		setRespContentType(ctx)
		return
	}

	save.saveKeyValue(key, val, ctx.Time(), delete)

}

func setRespContentType(ctx RequestCtx) {
	switch c := ctx.(type) {
	case *fasthttp.RequestCtx:
		c.Response.Header.SetCanonical(strContentType, strApplicationJSON)
	}
}

// Sum allows for the summation of all values for a given key using GET /metric/{key}/sum
func (km *KeysMap) Sum(ctx *fasthttp.RequestCtx) {
	sum(ctx, km)
}

// sum is the unit testable logic for the metric function
func sum(ctx *fasthttp.RequestCtx, doSum summer) {
	key := ctx.UserValue(routerKey)

	sum, err := doSum.getSum(key)
	if err != nil {
		ctx.Error(fmt.Sprintf("%s", err.Error()), http.StatusNotFound)
		setRespContentType(ctx)
		return
	}

	setRespContentType(ctx)
	ctx.Response.SetStatusCode(http.StatusOK)

	if err := json.NewEncoder(ctx).Encode(sum); err != nil {
		ctx.Error(err.Error(), fasthttp.StatusInternalServerError)
		setRespContentType(ctx)
	}

}
