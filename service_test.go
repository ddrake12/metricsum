package metricsum

import (
	"strconv"
	"testing"
	"time"

	"github.com/ddrake12/metricsum/mocks"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
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
			reqCtx.EXPECT().Error("invalid body JSON: unexpected end of JSON input", 400)
		} else if tt.invalidErr {
			reqCtx.EXPECT().Error("could not interpret value - ensure JSON is correct and that value does not equal zero to be interepreted", 400)
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
