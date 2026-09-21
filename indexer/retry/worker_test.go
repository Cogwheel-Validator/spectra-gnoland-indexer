package retry_test

import (
	"errors"
	"testing"
	"time"

	"github.com/Cogwheel-Validator/spectra-gnoland-indexer/indexer/retry"
	"github.com/stretchr/testify/assert"
)

func TestGenericRetryQuery_NotReadyThenErrorThenSuccess(t *testing.T) {
	calls := 0
	fn := func(args ...any) (int, error) {
		calls++
		switch calls {
		case 1, 2:
			return 0, retry.ErrNotReady
		case 3:
			return 0, errors.New("rpc down")
		}
		return 42, nil
	}

	res := <-retry.GenericRetryQuery(6, 3, time.Millisecond, time.Millisecond, fn)

	assert.True(t, res.Success)
	assert.Equal(t, 42, res.Value)
	assert.Equal(t, 4, calls)
}

func TestGenericRetryQuery_NotReadyExhausted(t *testing.T) {
	calls := 0
	fn := func(args ...any) (int, error) {
		calls++
		return 0, retry.ErrNotReady
	}

	res := <-retry.GenericRetryQuery(3, 3, time.Millisecond, time.Millisecond, fn)

	assert.False(t, res.Success)
	assert.ErrorIs(t, res.Error, retry.ErrNotReady)
	assert.Equal(t, 3, calls)
}
