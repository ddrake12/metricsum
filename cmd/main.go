package main

import (
	"log"

	"github.com/ddrake12/metricsum"
	"github.com/fasthttp/router"
	"github.com/valyala/fasthttp"
)

func main() {
	r := router.New()

	keysMap := metricsum.NewKeysMap()
	r.POST("/metric/{key}", keysMap.Metric)
	// r.GET("/metric/{key}/sum", metricsum.Sum)

	log.Fatal(fasthttp.ListenAndServe(":8080", r.Handler))
}
