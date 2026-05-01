/*
Copyright © 2025 Takayuki Nagatomi <tnagatomi@okweird.net>

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in
all copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
THE SOFTWARE.
*/
package api

import (
	"errors"
	"testing"
	"time"
)

func testRetryConfig() (retryConfig, *[]time.Duration) {
	var sleeps []time.Duration
	cfg := retryConfig{
		maxAttempts: 3,
		baseDelay:   10 * time.Millisecond,
		maxDelay:    40 * time.Millisecond,
		sleep:       func(d time.Duration) { sleeps = append(sleeps, d) },
		retryable:   isRetryable,
	}
	return cfg, &sleeps
}

func TestWithRetry_SuccessOnFirstAttempt(t *testing.T) {
	cfg, sleeps := testRetryConfig()
	calls := 0

	err := withRetry(func() error {
		calls++
		return nil
	}, cfg)

	if err != nil {
		t.Fatalf("withRetry() error = %v, want nil", err)
	}
	if calls != 1 {
		t.Errorf("calls = %d, want 1", calls)
	}
	if len(*sleeps) != 0 {
		t.Errorf("sleeps = %v, want none", *sleeps)
	}
}

func TestWithRetry_RetryableThenSuccess(t *testing.T) {
	cfg, sleeps := testRetryConfig()
	calls := 0

	err := withRetry(func() error {
		calls++
		if calls < 3 {
			return &TransientError{StatusCode: 502}
		}
		return nil
	}, cfg)

	if err != nil {
		t.Fatalf("withRetry() error = %v, want nil", err)
	}
	if calls != 3 {
		t.Errorf("calls = %d, want 3", calls)
	}
	wantSleeps := []time.Duration{10 * time.Millisecond, 20 * time.Millisecond}
	if len(*sleeps) != len(wantSleeps) {
		t.Fatalf("sleeps = %v, want %v", *sleeps, wantSleeps)
	}
	for i, d := range *sleeps {
		if d != wantSleeps[i] {
			t.Errorf("sleeps[%d] = %v, want %v", i, d, wantSleeps[i])
		}
	}
}

func TestWithRetry_NonRetryableErrorStops(t *testing.T) {
	cfg, sleeps := testRetryConfig()
	calls := 0
	wantErr := &NotFoundError{ResourceType: ResourceTypeRepository}

	err := withRetry(func() error {
		calls++
		return wantErr
	}, cfg)

	if !errors.Is(err, wantErr) {
		t.Errorf("withRetry() error = %v, want %v", err, wantErr)
	}
	if calls != 1 {
		t.Errorf("calls = %d, want 1 (non-retryable should not retry)", calls)
	}
	if len(*sleeps) != 0 {
		t.Errorf("sleeps = %v, want none", *sleeps)
	}
}

func TestWithRetry_AllAttemptsFail(t *testing.T) {
	cfg, sleeps := testRetryConfig()
	calls := 0
	wantErr := &RateLimitError{}

	err := withRetry(func() error {
		calls++
		return wantErr
	}, cfg)

	if !errors.Is(err, wantErr) {
		t.Errorf("withRetry() error = %v, want %v", err, wantErr)
	}
	if calls != cfg.maxAttempts {
		t.Errorf("calls = %d, want %d", calls, cfg.maxAttempts)
	}
	// Sleeps happen between attempts, so maxAttempts-1 sleeps total.
	if len(*sleeps) != cfg.maxAttempts-1 {
		t.Errorf("len(sleeps) = %d, want %d", len(*sleeps), cfg.maxAttempts-1)
	}
}

func TestWithRetry_BackoffCappedByMaxDelay(t *testing.T) {
	var sleeps []time.Duration
	cfg := retryConfig{
		maxAttempts: 5,
		baseDelay:   10 * time.Millisecond,
		maxDelay:    25 * time.Millisecond,
		sleep:       func(d time.Duration) { sleeps = append(sleeps, d) },
		retryable:   isRetryable,
	}

	_ = withRetry(func() error {
		return &TransientError{StatusCode: 503}
	}, cfg)

	want := []time.Duration{
		10 * time.Millisecond,
		20 * time.Millisecond,
		25 * time.Millisecond,
		25 * time.Millisecond,
	}
	if len(sleeps) != len(want) {
		t.Fatalf("sleeps = %v, want %v", sleeps, want)
	}
	for i, d := range sleeps {
		if d != want[i] {
			t.Errorf("sleeps[%d] = %v, want %v", i, d, want[i])
		}
	}
}

func TestWithRetry_CustomRetryablePredicate(t *testing.T) {
	cfg, sleeps := testRetryConfig()
	cfg.retryable = IsRateLimit
	calls := 0

	// TransientError is normally retryable but the custom predicate excludes
	// it; the call should run exactly once.
	err := withRetry(func() error {
		calls++
		return &TransientError{StatusCode: 503}
	}, cfg)

	if !IsTransient(err) {
		t.Errorf("withRetry() error = %v, want TransientError", err)
	}
	if calls != 1 {
		t.Errorf("calls = %d, want 1 (transient must not retry under rate-limit-only predicate)", calls)
	}
	if len(*sleeps) != 0 {
		t.Errorf("sleeps = %v, want none", *sleeps)
	}
}

func TestWithRetry_RateLimitOnlyStillRetriesRateLimit(t *testing.T) {
	cfg, _ := testRetryConfig()
	cfg.retryable = IsRateLimit
	calls := 0

	err := withRetry(func() error {
		calls++
		if calls < 2 {
			return &RateLimitError{}
		}
		return nil
	}, cfg)

	if err != nil {
		t.Errorf("withRetry() error = %v, want nil", err)
	}
	if calls != 2 {
		t.Errorf("calls = %d, want 2", calls)
	}
}

func TestIsRetryable(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want bool
	}{
		{"RateLimitError", &RateLimitError{}, true},
		{"TransientError", &TransientError{StatusCode: 502}, true},
		{"NotFoundError", &NotFoundError{ResourceType: ResourceTypeRepository}, false},
		{"ForbiddenError", &ForbiddenError{}, false},
		{"plain error", errors.New("oops"), false},
		{"nil", nil, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isRetryable(tt.err); got != tt.want {
				t.Errorf("isRetryable(%v) = %v, want %v", tt.err, got, tt.want)
			}
		})
	}
}
