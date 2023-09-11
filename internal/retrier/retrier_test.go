package retrier

import (
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestRetrier(t *testing.T) {
	type args struct {
		err         error
		retries     []time.Duration
		shouldRetry ShouldRetry
	}
	tests := []struct {
		name         string
		args         args
		wantAttempts int
	}{
		{
			name:         "No retry",
			wantAttempts: 1,
			args:         args{err: errors.New("fake error"), retries: []time.Duration{}, shouldRetry: AlwaysRetry},
		},
		{
			name:         "No error",
			wantAttempts: 1,
			args:         args{err: nil, retries: []time.Duration{time.Millisecond, time.Millisecond}, shouldRetry: AlwaysRetry},
		},
		{
			name:         "N times",
			wantAttempts: 3,
			args:         args{err: errors.New("fake error"), retries: []time.Duration{time.Millisecond, time.Millisecond}, shouldRetry: AlwaysRetry},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &Retrier{
				attempt:     0,
				retries:     tt.args.retries,
				shouldRetry: tt.args.shouldRetry,
			}
			attemptCount := 0
			var err error
			for r.Next(err) {
				err = tt.args.err
				attemptCount += 1
			}
			assert.Equal(t, tt.wantAttempts, attemptCount)
		})
	}
}
