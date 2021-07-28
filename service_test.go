package metricsum

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"testing"
	"time"

	"github.com/ddrake12/metricsum/mocks"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/valyala/fasthttp"
)

// testSaver is used to make sure saveKeyValue is called
type testSaver struct {
	key     interface{}
	value   int
	reqTime time.Time
}

func (ts *testSaver) saveKeyValue(key interface{}, value int, reqTime time.Time, delete deleter) {
	ts.key = key
	ts.value = value
	ts.reqTime = reqTime
}

func Test_metric(t *testing.T) {

	testValStr := strconv.Itoa(testValue)
	tests := []struct {
		name       string
		postBody   []uint8
		malformErr bool
		invalidErr bool
	}{
		{"test valid input", []byte(`{"value":` + testValStr + `}`), false, false},
		{"test malformed JSON input", []byte(`{"value:3`), true, false},
		{"test invalid JSON input", []byte(`{"val":3}`), false, true},
	}
	for _, tt := range tests {
		now := time.Now()

		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()
		reqCtx := mocks.NewMockRequestCtx(mockCtrl)
		reqCtx.EXPECT().UserValue(routerKey).Return(testKey)
		reqCtx.EXPECT().PostBody().Return(tt.postBody)

		if tt.malformErr {
			reqCtx.EXPECT().Error("invalid body JSON: unexpected end of JSON input", http.StatusBadRequest)
		} else if tt.invalidErr {
			reqCtx.EXPECT().Error("could not interpret value - ensure JSON is correct and that value does not round to zero to be interepreted", http.StatusBadRequest)
		} else {
			reqCtx.EXPECT().Time().Return(now)
		}

		tSaver := &testSaver{}

		t.Run(tt.name, func(t *testing.T) {
			metric(reqCtx, tSaver, &KeysMap{})
			if !tt.malformErr && !tt.invalidErr {
				assert.Equal(t, tSaver.key, testKey)
				assert.Equal(t, tSaver.value, testValue)
				assert.Equal(t, tSaver.reqTime, now)
			}
		})
	}
}

// testSummer is used to mock the call to getSum
type testSummer struct {
	wantErr bool
}

func (ts *testSummer) getSum(key interface{}) (int, error) {
	if ts.wantErr {
		return 0, errors.New(testErrorStr)
	}
	return testSum, nil
}

type ctxError struct {
	err string
}

func Test_sum(t *testing.T) {

	// TODO: discuss in code review if mock style is preferred or using the 3rd party object like this

	tests := []struct {
		name    string
		wantErr bool
	}{
		{"test valid sum", false},
		{"test invalid sum", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			tSummer := &testSummer{tt.wantErr}
			ctx := &fasthttp.RequestCtx{}
			ctx.SetUserValue(routerKey, testKey)

			sum(ctx, tSummer)

			if tt.wantErr {
				assert.Contains(t, string(ctx.Response.Body()), testErrorStr)
			} else {
				var sum int
				err := json.Unmarshal(ctx.Response.Body(), &sum)
				assert.NoError(t, err)
				assert.Equal(t, testSum, sum)
			}
		})
	}
}
