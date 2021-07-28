package metricsum

import (
	"bytes"
	"fmt"
	"log"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

var (
	testKey     = "testKey"
	diffKey     = "diffKey"
	testValue   = 1
	diffValue   = 2
	testCounter = 2
)

// mockTimeAfter will stub out time after to return immediately. Also tests that given value is always 1 hour.
func mockTimeAfter(t *testing.T) {
	tChan := make(chan time.Time, 1)
	tChan <- time.Now() // immediately return when called
	timeAfter = func(d time.Duration) <-chan time.Time {
		assert.Equal(t, 1*time.Hour, d)
		return tChan
	}

}

// mockLogger sets output to a *bytes.Buffer and returns it to check needed output
func mockLogger(km *KeysMap) *bytes.Buffer {
	var buf bytes.Buffer
	km.logger.SetOutput(&buf)
	return &buf
}

// testDeleter is used to make sure startDeleteTimer is called
type testDeleter struct {
	key     interface{}
	value   int
	reqTime time.Time
	done    chan struct{}
}

func newTestDeleter() *testDeleter {
	done := make(chan struct{})
	return &testDeleter{
		done: done,
	}
}

func (td *testDeleter) startDeleteTimer(key interface{}, value int, reqTime time.Time) {
	td.key = key
	td.value = value
	td.reqTime = reqTime
	td.done <- struct{}{}
}

func TestNewKeysMap(t *testing.T) {
	tests := []struct {
		name string
		want *KeysMap
	}{
		{"test valid constructor", &KeysMap{&sync.Mutex{}, make(map[interface{}]*values), log.Default()}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NewKeysMap()
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestKeysMap_saveKeyValue(t *testing.T) {

	sameKeySameVal := "test same key for same value "
	sameKeyDiffVal := "test same key for different value"
	diffKeysSameVal := "test different keys for same value"
	diffKeysDiffVals := "test different keys for different values"
	deleteCalledProperly := "test delete is called properly"

	tests := []struct {
		name string
	}{
		{sameKeySameVal},
		{sameKeyDiffVal},
		{diffKeysSameVal},
		{diffKeysDiffVals},
		{deleteCalledProperly},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			km := NewKeysMap()

			switch tt.name {
			case sameKeySameVal:
				testSameKeySameValue(t, km)

			case sameKeyDiffVal:
				testSameKeyDiffValue(t, km)

			case diffKeysSameVal:
				testDiffKeySameValue(t, km)

			case diffKeysDiffVals:
				testDiffKeyDiffValue(t, km)
			case deleteCalledProperly:
				testDeleteCalledProperly(t, km)
			}

		})
	}
}

func testSameKeySameValue(t *testing.T, km *KeysMap) {

	km.saveKeyValue(testKey, testValue, time.Now(), km)
	km.saveKeyValue(testKey, testValue, time.Now(), km)

	expectedCount := 2
	expectedNumVals := 1

	checkKeyAndValExpectations(t, km, testKey, testValue, expectedCount, expectedNumVals)

}

func testSameKeyDiffValue(t *testing.T, km *KeysMap) {

	km.saveKeyValue(testKey, testValue, time.Now(), km)
	km.saveKeyValue(testKey, diffValue, time.Now(), km)

	expectedNumVals := 2
	expectedCount := 1
	checkKeyAndValExpectations(t, km, testKey, testValue, expectedCount, expectedNumVals)
	checkKeyAndValExpectations(t, km, testKey, diffValue, expectedCount, expectedNumVals)
}

func testDiffKeySameValue(t *testing.T, km *KeysMap) {
	km.saveKeyValue(testKey, testValue, time.Now(), km)
	km.saveKeyValue(diffKey, testValue, time.Now(), km)

	expectedNumVals := 1
	expectedCount := 1
	checkKeyAndValExpectations(t, km, testKey, testValue, expectedCount, expectedNumVals)
	checkKeyAndValExpectations(t, km, diffKey, testValue, expectedCount, expectedNumVals)

}

func testDiffKeyDiffValue(t *testing.T, km *KeysMap) {
	km.saveKeyValue(testKey, testValue, time.Now(), km)
	km.saveKeyValue(diffKey, diffValue, time.Now(), km)

	expectedNumVals := 1
	expectedCount := 1
	checkKeyAndValExpectations(t, km, testKey, testValue, expectedCount, expectedNumVals)
	checkKeyAndValExpectations(t, km, diffKey, diffValue, expectedCount, expectedNumVals)
}

func checkKeyAndValExpectations(t *testing.T, km *KeysMap, key string, expectedValue, expectedCount, expectedNumVals int) {
	vals, ok := km.keyToValues[key]
	assert.True(t, ok)

	found := false
	numVals := 0
	for val, counter := range vals.valueToCounter {
		if val == expectedValue {
			found = true
			assert.Equal(t, expectedCount, counter)
		}
		numVals++
	}
	assert.Equal(t, expectedNumVals, numVals)
	assert.True(t, found)
}

func testDeleteCalledProperly(t *testing.T, km *KeysMap) {
	mockTimeAfter(t)

	tDeleter := newTestDeleter()
	now := time.Now()

	km.saveKeyValue(testKey, testValue, now, tDeleter)
	<-tDeleter.done

	assert.Equal(t, testKey, tDeleter.key)
	assert.Equal(t, testValue, tDeleter.value)
	assert.Equal(t, now, tDeleter.reqTime)

}

func Test_KeysMap_startDeleteTimer(t *testing.T) {

	tests := []struct {
		name      string
		keyMapErr bool
		valMapErr bool
	}{
		{"test delete", false, false},
		{"test non existent key in KeysMap error", true, false},
		{"test non existent key in valueToCounter map error", false, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			km := NewKeysMap()

			mockTimeAfter(t)
			buf := mockLogger(km)

			if !tt.keyMapErr && !tt.valMapErr {
				testMap := make(map[int]int)
				testMap[testValue] = testCounter
				km.keyToValues[testKey] = &values{testMap, &sync.Mutex{}}
			} else if tt.valMapErr {
				testMap := make(map[int]int)
				km.keyToValues[testKey] = &values{testMap, &sync.Mutex{}}
			}

			reqTime := time.Now()
			km.startDeleteTimer(testKey, testValue, reqTime)

			if tt.keyMapErr {
				msg := fmt.Sprintf("was not able to locate the values for key: %v to decrement count after an hour? Investigate how this is possible. Value: %v ReqTime: %v", testKey, testValue, reqTime)
				assert.Contains(t, buf.String(), msg)
			} else if tt.valMapErr {
				msg := fmt.Sprintf("was not able to locate the counter for value: %v to decrement count after an hour? Investigate how this is possible. Key: %v ReqTime: %v", testValue, testKey, reqTime)
				assert.Contains(t, buf.String(), msg)
			} else {
				values := km.keyToValues[testKey]
				assert.Equal(t, testCounter-1, values.valueToCounter[testValue])
			}
		})
	}
}
