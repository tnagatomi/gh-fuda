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
	"github.com/cli/go-gh/v2/pkg/api"
)

// GraphQLAPI implements APIClient using GitHub GraphQL API
type GraphQLAPI struct {
	client *api.GraphQLClient
	// Cache for repository IDs to avoid redundant queries
	repoIDCache map[string]string
}

// NewGraphQLAPI creates a new GraphQL API client
func NewGraphQLAPI() (*GraphQLAPI, error) {
	client, err := api.NewGraphQLClient(api.ClientOptions{})
	if err != nil {
		return nil, err
	}
	return &GraphQLAPI{
		client:      client,
		repoIDCache: make(map[string]string),
	}, nil
}
