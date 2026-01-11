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
package executor

import (
	"bytes"
	"errors"
	"sync/atomic"
	"testing"
	"time"
)

func TestWorkerPool_Run_OrderPreserved(t *testing.T) {
	var buf bytes.Buffer
	wp := NewWorkerPool(&buf)

	jobs := make([]Job, 10)
	for i := 0; i < 10; i++ {
		id := i
		jobs[i] = Job{
			ID: id,
			Func: func() *JobResult {
				return &JobResult{
					ID:      id,
					Output:  "result",
					Success: true,
				}
			},
		}
	}

	results := wp.Run(jobs)

	if len(results) != 10 {
		t.Fatalf("expected 10 results, got %d", len(results))
	}

	// Verify results are in order
	for i, result := range results {
		if result.ID != i {
			t.Errorf("result[%d].ID = %d, want %d", i, result.ID, i)
		}
	}
}

func TestWorkerPool_Run_ConcurrencyLimit(t *testing.T) {
	var buf bytes.Buffer
	wp := NewWorkerPool(&buf)

	var concurrent int32
	var maxConcurrent int32

	jobs := make([]Job, 20)
	for i := 0; i < 20; i++ {
		jobs[i] = Job{
			ID: i,
			Func: func() *JobResult {
				// Track concurrent executions
				current := atomic.AddInt32(&concurrent, 1)

				// Update max if current is higher
				for {
					max := atomic.LoadInt32(&maxConcurrent)
					if current <= max {
						break
					}
					if atomic.CompareAndSwapInt32(&maxConcurrent, max, current) {
						break
					}
				}

				// Simulate work
				time.Sleep(10 * time.Millisecond)

				atomic.AddInt32(&concurrent, -1)
				return &JobResult{Success: true}
			},
		}
	}

	wp.Run(jobs)

	max := atomic.LoadInt32(&maxConcurrent)
	if max > int32(WorkerPoolSize) {
		t.Errorf("max concurrent = %d, want <= %d", max, WorkerPoolSize)
	}
}

func TestWorkerPool_Run_ErrorHandling(t *testing.T) {
	var buf bytes.Buffer
	wp := NewWorkerPool(&buf)

	expectedErr := errors.New("test error")

	jobs := []Job{
		{
			ID: 0,
			Func: func() *JobResult {
				return &JobResult{Success: true}
			},
		},
		{
			ID: 1,
			Func: func() *JobResult {
				return &JobResult{Success: false, Errors: []error{expectedErr}}
			},
		},
		{
			ID: 2,
			Func: func() *JobResult {
				return &JobResult{Success: true}
			},
		},
	}

	results := wp.Run(jobs)

	if len(results) != 3 {
		t.Fatalf("expected 3 results, got %d", len(results))
	}

	if !results[0].Success {
		t.Error("results[0] should be success")
	}
	if results[1].Success {
		t.Error("results[1] should be failure")
	}
	if len(results[1].Errors) != 1 || !errors.Is(results[1].Errors[0], expectedErr) {
		t.Errorf("results[1].Errors = %v, want [%v]", results[1].Errors, expectedErr)
	}
	if !results[2].Success {
		t.Error("results[2] should be success")
	}
}

func TestWorkerPool_Run_ProgressOutput(t *testing.T) {
	var buf bytes.Buffer
	wp := NewWorkerPool(&buf)

	jobs := []Job{
		{ID: 0, Func: func() *JobResult { return &JobResult{Success: true} }},
		{ID: 1, Func: func() *JobResult { return &JobResult{Success: true} }},
		{ID: 2, Func: func() *JobResult { return &JobResult{Success: true} }},
	}

	wp.Run(jobs)

	output := buf.String()
	// Progress output should contain completion counts
	if len(output) == 0 {
		t.Error("expected progress output, got empty string")
	}
}

func TestWorkerPool_Run_EmptyJobs(t *testing.T) {
	var buf bytes.Buffer
	wp := NewWorkerPool(&buf)

	results := wp.Run([]Job{})

	if len(results) != 0 {
		t.Errorf("expected 0 results, got %d", len(results))
	}
}