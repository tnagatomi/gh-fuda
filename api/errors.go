/*
Copyright © 2025 Takayuki Nagatomi <tommyt6073@gmail.com>

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
	"fmt"
	"strings"

	"github.com/google/go-github/v59/github"
)

// UnauthorizedError indicates 401 Unauthorized response
type UnauthorizedError struct{}

func (e *UnauthorizedError) Error() string {
	return "unauthorized"
}

// ForbiddenError indicates 403 Forbidden response
type ForbiddenError struct{}

func (e *ForbiddenError) Error() string {
	return "forbidden"
}

// NotFoundError indicates 404 Not Found response
type NotFoundError struct {
	Resource string
}

func (e *NotFoundError) Error() string {
	if e.Resource == "" {
		return "not found"
	}
	return fmt.Sprintf("%s not found", e.Resource)
}

// RateLimitError indicates 429 Too Many Requests response
type RateLimitError struct{}

func (e *RateLimitError) Error() string {
	return "rate limit exceeded"
}

// AlreadyExistsError indicates resource already exists
type AlreadyExistsError struct {
	Resource string
}

func (e *AlreadyExistsError) Error() string {
	if e.Resource == "" {
		return "already exists"
	}
	return fmt.Sprintf("%s already exists", e.Resource)
}

// Helper functions to check error types
func IsNotFound(err error) bool {
	var notFoundErr *NotFoundError
	return errors.As(err, &notFoundErr)
}

func IsUnauthorized(err error) bool {
	var unauthorizedErr *UnauthorizedError
	return errors.As(err, &unauthorizedErr)
}

func IsForbidden(err error) bool {
	var forbiddenErr *ForbiddenError
	return errors.As(err, &forbiddenErr)
}

func IsRateLimit(err error) bool {
	var rateLimitErr *RateLimitError
	return errors.As(err, &rateLimitErr)
}

func IsAlreadyExists(err error) bool {
	var alreadyExistsErr *AlreadyExistsError
	return errors.As(err, &alreadyExistsErr)
}

// wrapGitHubError converts GitHub API errors to our custom error types
func wrapGitHubError(err error, resource string) error {
	if err == nil {
		return nil
	}

	var ghErr *github.ErrorResponse
	if errors.As(err, &ghErr) {
		switch ghErr.Response.StatusCode {
		case 401:
			return &UnauthorizedError{}
		case 403:
			return &ForbiddenError{}
		case 404:
			return &NotFoundError{Resource: resource}
		case 429:
			return &RateLimitError{}
		case 422:
			if strings.Contains(err.Error(), "already_exists") {
				return &AlreadyExistsError{Resource: resource}
			}
			return fmt.Errorf("GitHub API error (status %d): %s", ghErr.Response.StatusCode, ghErr.Message)
		default:
			// Return the original error with status code for other cases
			return fmt.Errorf("GitHub API error (status %d): %s", ghErr.Response.StatusCode, ghErr.Message)
		}
	}

	// If it's not a GitHub error response, return the original error
	return err
}
