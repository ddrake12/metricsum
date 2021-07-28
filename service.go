package metricsum

//go:generate mockgen -source=service.go -destination=mocks/RequestCtx.go -package=mocks

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/valyala/fasthttp"
)

const routerKey = "key"

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

// Value allows for easy unmarshalling of POST /metric/{key} body data
type Value struct {
	Value int `json:"value"`
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
		ctx.Error(fmt.Sprintf("invalid body JSON: %v", err), 400)
		return
	}

	if value.Value == 0 {
		ctx.Error("could not interpret value - ensure JSON is correct and that value does not equal zero to be interepreted", 400)
		return
	}

	save.saveKeyValue(key, value.Value, ctx.Time(), delete)

}
