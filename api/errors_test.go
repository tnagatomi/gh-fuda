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
	"testing"
)

func TestHelperFunctions(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		checkFn  func(error) bool
		expected bool
	}{
		{
			name:     "IsNotFound with NotFoundError",
			err:      &NotFoundError{ResourceType: ResourceTypeLabel},
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
			err:      &AlreadyExistsError{ResourceType: ResourceTypeLabel},
			checkFn:  IsRateLimit,
			expected: false,
		},
		{
			name:     "IsAlreadyExists with AlreadyExistsError",
			err:      &AlreadyExistsError{ResourceType: ResourceTypeLabel},
			checkFn:  IsAlreadyExists,
			expected: true,
		},
		{
			name:     "IsAlreadyExists with other error",
			err:      &NotFoundError{ResourceType: ResourceTypeLabel},
			checkFn:  IsAlreadyExists,
			expected: false,
		},
		{
			name:     "IsScopeError with ScopeError",
			err:      &ScopeError{RequiredScope: "write:discussion"},
			checkFn:  IsScopeError,
			expected: true,
		},
		{
			name:     "IsScopeError with other error",
			err:      &ForbiddenError{},
			checkFn:  IsScopeError,
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
			name:    "NotFoundError with resource type repository",
			err:     &NotFoundError{ResourceType: ResourceTypeRepository},
			wantMsg: "repository not found",
		},
		{
			name:    "NotFoundError with resource type label",
			err:     &NotFoundError{ResourceType: ResourceTypeLabel},
			wantMsg: "label not found",
		},
		{
			name:    "NotFoundError without resource type",
			err:     &NotFoundError{ResourceType: ""},
			wantMsg: "not found",
		},
		{
			name:    "RateLimitError",
			err:     &RateLimitError{},
			wantMsg: "rate limit exceeded",
		},
		{
			name:    "AlreadyExistsError with resource type label",
			err:     &AlreadyExistsError{ResourceType: ResourceTypeLabel},
			wantMsg: "label already exists",
		},
		{
			name:    "AlreadyExistsError without resource type",
			err:     &AlreadyExistsError{ResourceType: ""},
			wantMsg: "already exists",
		},
		{
			name:    "ScopeError with required scope",
			err:     &ScopeError{RequiredScope: "write:discussion"},
			wantMsg: "insufficient token scopes (required: write:discussion)",
		},
		{
			name:    "ScopeError without required scope",
			err:     &ScopeError{RequiredScope: ""},
			wantMsg: "insufficient token scopes",
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
