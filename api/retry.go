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

import "time"

// retryConfig controls the behavior of withRetry.
type retryConfig struct {
	maxAttempts int
	baseDelay   time.Duration
	maxDelay    time.Duration
	sleep       func(time.Duration)
	// retryable returns true for errors that should be retried. If nil,
	// isRetryable is used (rate limit + transient).
	retryable func(error) bool
}

func defaultRetryConfig() retryConfig {
	return retryConfig{
		maxAttempts: 3,
		baseDelay:   1 * time.Second,
		maxDelay:    8 * time.Second,
		sleep:       time.Sleep,
		retryable:   isRetryable,
	}
}

// withRetry runs fn and retries on retryable errors using exponential backoff.
// It returns the last error if all attempts fail or fn returns a non-retryable error.
func withRetry(fn func() error, cfg retryConfig) error {
	var err error
	delay := cfg.baseDelay
	for attempt := 1; attempt <= cfg.maxAttempts; attempt++ {
		err = fn()
		if err == nil {
			return nil
		}
		if !cfg.retryable(err) || attempt == cfg.maxAttempts {
			return err
		}
		cfg.sleep(delay)
		delay *= 2
		if delay > cfg.maxDelay {
			delay = cfg.maxDelay
		}
	}
	return err
}

// isRetryable returns true for errors that may succeed on retry: rate limits
// and transient (5xx, network) failures.
func isRetryable(err error) bool {
	return IsRateLimit(err) || IsTransient(err)
}
