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
	"net/http"
	"testing"

	"github.com/google/go-github/v59/github"
)

func TestWrapGitHubError(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		resource string
		want     error
		wantMsg  string
	}{
		{
			name:     "nil error",
			err:      nil,
			resource: "test resource",
			want:     nil,
			wantMsg:  "",
		},
		{
			name: "401 unauthorized",
			err: &github.ErrorResponse{
				Response: &http.Response{StatusCode: 401},
				Message:  "Bad credentials",
			},
			resource: "test resource",
			want:     &UnauthorizedError{},
			wantMsg:  "unauthorized",
		},
		{
			name: "403 forbidden",
			err: &github.ErrorResponse{
				Response: &http.Response{StatusCode: 403},
				Message:  "Resource not accessible",
			},
			resource: "test resource",
			want:     &ForbiddenError{},
			wantMsg:  "forbidden",
		},
		{
			name: "404 not found with resource",
			err: &github.ErrorResponse{
				Response: &http.Response{StatusCode: 404},
				Message:  "Not Found",
			},
			resource: "repository \"owner/repo\"",
			want:     &NotFoundError{Resource: "repository \"owner/repo\""},
			wantMsg:  "repository \"owner/repo\" not found",
		},
		{
			name: "429 rate limit",
			err: &github.ErrorResponse{
				Response: &http.Response{StatusCode: 429},
				Message:  "API rate limit exceeded",
			},
			resource: "test resource",
			want:     &RateLimitError{},
			wantMsg:  "rate limit exceeded",
		},
		{
			name: "422 already exists",
			err: &github.ErrorResponse{
				Response: &http.Response{StatusCode: 422},
				Message:  "Validation Failed",
				Errors: []github.Error{
					{Code: "already_exists"},
				},
			},
			resource: "label \"bug\" on \"owner/repo\"",
			want:     &AlreadyExistsError{Resource: "label \"bug\" on \"owner/repo\""},
			wantMsg:  "label \"bug\" on \"owner/repo\" already exists",
		},
		{
			name: "422 other validation error",
			err: &github.ErrorResponse{
				Response: &http.Response{StatusCode: 422},
				Message:  "Validation Failed",
				Errors: []github.Error{
					{Code: "invalid"},
				},
			},
			resource: "test resource",
			want:     fmt.Errorf("GitHub API error (status 422): Validation Failed"),
			wantMsg:  "GitHub API error (status 422): Validation Failed",
		},
		{
			name: "500 server error",
			err: &github.ErrorResponse{
				Response: &http.Response{StatusCode: 500},
				Message:  "Internal Server Error",
			},
			resource: "test resource",
			want:     fmt.Errorf("GitHub API error (status 500): Internal Server Error"),
			wantMsg:  "GitHub API error (status 500): Internal Server Error",
		},
		{
			name:     "non-GitHub error",
			err:      errors.New("network error"),
			resource: "test resource",
			want:     errors.New("network error"),
			wantMsg:  "network error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := wrapGitHubError(tt.err, tt.resource)

			// Check if error is nil
			if tt.want == nil {
				if got != nil {
					t.Errorf("wrapGitHubError() = %v, want nil", got)
				}
				return
			}

			// Check error message
			if got.Error() != tt.wantMsg {
				t.Errorf("wrapGitHubError() error message = %v, want %v", got.Error(), tt.wantMsg)
			}

			// Check error type for custom errors
			switch want := tt.want.(type) {
			case *UnauthorizedError:
				if _, ok := got.(*UnauthorizedError); !ok {
					t.Errorf("wrapGitHubError() = %T, want %T", got, want)
				}
			case *ForbiddenError:
				if _, ok := got.(*ForbiddenError); !ok {
					t.Errorf("wrapGitHubError() = %T, want %T", got, want)
				}
			case *NotFoundError:
				gotErr, ok := got.(*NotFoundError)
				if !ok {
					t.Errorf("wrapGitHubError() = %T, want %T", got, want)
				} else if gotErr.Resource != want.Resource {
					t.Errorf("NotFoundError.Resource = %v, want %v", gotErr.Resource, want.Resource)
				}
			case *RateLimitError:
				if _, ok := got.(*RateLimitError); !ok {
					t.Errorf("wrapGitHubError() = %T, want %T", got, want)
				}
			case *AlreadyExistsError:
				gotErr, ok := got.(*AlreadyExistsError)
				if !ok {
					t.Errorf("wrapGitHubError() = %T, want %T", got, want)
				} else if gotErr.Resource != want.Resource {
					t.Errorf("AlreadyExistsError.resource = %v, want %v", gotErr.Resource, want.Resource)
				}
			}
		})
	}
}

func TestHelperFunctions(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		checkFn  func(error) bool
		expected bool
	}{
		{
			name:     "IsNotFound with NotFoundError",
			err:      &NotFoundError{Resource: "test"},
			checkFn:  IsNotFound,
			expected: true,
		},
		{
			name:     "IsNotFound with other error",
			err:      &UnauthorizedError{},
			checkFn:  IsNotFound,
			expected: false,
		},
		{
			name:     "IsUnauthorized with UnauthorizedError",
			err:      &UnauthorizedError{},
			checkFn:  IsUnauthorized,
			expected: true,
		},
		{
			name:     "IsUnauthorized with other error",
			err:      &ForbiddenError{},
			checkFn:  IsUnauthorized,
			expected: false,
		},
		{
			name:     "IsForbidden with ForbiddenError",
			err:      &ForbiddenError{},
			checkFn:  IsForbidden,
			expected: true,
		},
		{
			name:     "IsForbidden with other error",
			err:      &RateLimitError{},
			checkFn:  IsForbidden,
			expected: false,
		},
		{
			name:     "IsRateLimit with RateLimitError",
			err:      &RateLimitError{},
			checkFn:  IsRateLimit,
			expected: true,
		},
		{
			name:     "IsRateLimit with other error",
			err:      &AlreadyExistsError{Resource: "test"},
			checkFn:  IsRateLimit,
			expected: false,
		},
		{
			name:     "IsAlreadyExists with AlreadyExistsError",
			err:      &AlreadyExistsError{Resource: "test"},
			checkFn:  IsAlreadyExists,
			expected: true,
		},
		{
			name:     "IsAlreadyExists with other error",
			err:      &NotFoundError{Resource: "test"},
			checkFn:  IsAlreadyExists,
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.checkFn(tt.err); got != tt.expected {
				t.Errorf("%s() = %v, want %v", tt.name, got, tt.expected)
			}
		})
	}
}

func TestErrorMessages(t *testing.T) {
	tests := []struct {
		name    string
		err     error
		wantMsg string
	}{
		{
			name:    "UnauthorizedError",
			err:     &UnauthorizedError{},
			wantMsg: "unauthorized",
		},
		{
			name:    "ForbiddenError",
			err:     &ForbiddenError{},
			wantMsg: "forbidden",
		},
		{
			name:    "NotFoundError with resource",
			err:     &NotFoundError{Resource: "label \"bug\" on \"owner/repo\""},
			wantMsg: "label \"bug\" on \"owner/repo\" not found",
		},
		{
			name:    "NotFoundError without resource",
			err:     &NotFoundError{Resource: ""},
			wantMsg: "not found",
		},
		{
			name:    "RateLimitError",
			err:     &RateLimitError{},
			wantMsg: "rate limit exceeded",
		},
		{
			name:    "AlreadyExistsError with resource",
			err:     &AlreadyExistsError{Resource: "label \"bug\" on \"owner/repo\""},
			wantMsg: "label \"bug\" on \"owner/repo\" already exists",
		},
		{
			name:    "AlreadyExistsError without resource",
			err:     &AlreadyExistsError{Resource: ""},
			wantMsg: "already exists",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.err.Error(); got != tt.wantMsg {
				t.Errorf("Error() = %v, want %v", got, tt.wantMsg)
			}
		})
	}
}
