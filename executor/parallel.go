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
	"fmt"
	"io"
	"sync"
)

const (
	// WorkerPoolSize is the number of concurrent workers for parallel operations
	WorkerPoolSize = 5
)

// Job represents a unit of work to be executed by a worker
type Job struct {
	ID   int
	Func func() *JobResult
}

// JobResult represents the result of a job execution
type JobResult struct {
	ID      int
	Output  string
	Success bool
	Error   error
}

// WorkerPool manages parallel job execution with a fixed number of workers
type WorkerPool struct {
	workers    int
	jobs       chan Job
	results    chan *JobResult
	wg         sync.WaitGroup
	out        io.Writer
	totalJobs  int
	completed  int
	mu         sync.Mutex
}

// NewWorkerPool creates a new worker pool with the specified number of workers
func NewWorkerPool(out io.Writer) *WorkerPool {
	return &WorkerPool{
		workers: WorkerPoolSize,
		out:     out,
	}
}

// Run executes all jobs in parallel and returns results in order
func (wp *WorkerPool) Run(jobs []Job) []*JobResult {
	wp.totalJobs = len(jobs)
	wp.completed = 0
	wp.jobs = make(chan Job, len(jobs))
	wp.results = make(chan *JobResult, len(jobs))

	// Start workers
	for i := 0; i < wp.workers; i++ {
		wp.wg.Add(1)
		go wp.worker()
	}

	// Send jobs to workers
	for _, job := range jobs {
		wp.jobs <- job
	}
	close(wp.jobs)

	// Wait for all workers to complete
	wp.wg.Wait()
	close(wp.results)

	// Collect results and sort by ID to maintain order
	resultsMap := make(map[int]*JobResult)
	for result := range wp.results {
		resultsMap[result.ID] = result
	}

	// Return results in order
	orderedResults := make([]*JobResult, len(jobs))
	for i := range jobs {
		orderedResults[i] = resultsMap[i]
	}

	return orderedResults
}

// worker processes jobs from the jobs channel
func (wp *WorkerPool) worker() {
	defer wp.wg.Done()
	for job := range wp.jobs {
		result := job.Func()
		result.ID = job.ID

		wp.mu.Lock()
		wp.completed++
		_, _ = fmt.Fprintf(wp.out, "\rProgress: %d/%d completed", wp.completed, wp.totalJobs)
		wp.mu.Unlock()

		wp.results <- result
	}
}

// ClearProgress clears the progress line and moves to a new line
func (wp *WorkerPool) ClearProgress() {
	_, _ = fmt.Fprintf(wp.out, "\r\033[K") // Clear the line
}
