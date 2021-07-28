package metricsum

import (
	"log"
	"sync"
	"time"
)

// set timeAfter in tests to control when it returns
var (
	timeAfter = time.After
)

// use deleter to mock the call to startDeleteTimer
type deleter interface {
	startDeleteTimer(interface{}, int, time.Time)
}

// KeysMap has methods Metric and Sum that can be used as a fasthttp.RequestHandler for routing with fasthttp/routing.
// All access to the internal keys map is mutex protected and thread safe.
type KeysMap struct {
	keyMu       *sync.Mutex
	keyToValues map[interface{}]*values
	logger      *log.Logger // TODO: implement a real logging framework
}

// NewKeysMap initializes a new KeysMap
func NewKeysMap() *KeysMap {
	return &KeysMap{
		keyMu:       &sync.Mutex{},
		keyToValues: make(map[interface{}]*values),
		logger:      log.Default(),
	}
}

// values contains a list of all values for a given key. The mutex must be locked before reading/writing to the value/counter
type values struct {
	valueToCounter map[int]int
	valueMu        *sync.Mutex
}

// saveKeyValue records the value for the given key. Key is kept as an interface from fasthttp since type assertion isn't necessary
func (km *KeysMap) saveKeyValue(key interface{}, value int, reqTime time.Time, delete deleter) {
	km.keyMu.Lock()
	vals, ok := km.keyToValues[key]

	if !ok {
		// have to hold the key mutex to ensure the same key isn't created at the same time
		valueMu := &sync.Mutex{}
		valueToCounter := make(map[int]int)
		valueToCounter[value] = 1
		vals = &values{
			valueToCounter: valueToCounter,
			valueMu:        valueMu,
		}

		km.keyToValues[key] = vals
		km.keyMu.Unlock()

	} else {

		km.keyMu.Unlock()

		vals.valueMu.Lock()

		count, ok := vals.valueToCounter[value]
		if !ok {
			vals.valueToCounter[value] = 1
		} else {
			vals.valueToCounter[value] = count + 1
		}

		vals.valueMu.Unlock()

	}

	go delete.startDeleteTimer(key, value, reqTime)

	return
}

func (km *KeysMap) startDeleteTimer(key interface{}, value int, reqTime time.Time) {

	expires := reqTime.Add(1 * time.Hour).Sub(reqTime) // one hour after req time

	// simple stub
	<-timeAfter(expires)

	km.keyMu.Lock()
	vals, ok := km.keyToValues[key]
	km.keyMu.Unlock()

	if !ok {
		km.logger.Printf("was not able to locate the values for key: %v to decrement count after an hour? Investigate how this is possible. Value: %v ReqTime: %v", key, value, reqTime)
	} else {

		vals.valueMu.Lock()

		counter, ok := vals.valueToCounter[value]
		if !ok {
			km.logger.Printf("was not able to locate the counter for value: %v to decrement count after an hour? Investigate how this is possible. Key: %v ReqTime: %v", value, key, reqTime)
		} else {
			vals.valueToCounter[value] = counter - 1
		}
		vals.valueMu.Unlock()
	}
}
