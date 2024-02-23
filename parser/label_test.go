/*
Copyright © 2024 Takayuki Nagatomi <tommyt6073@gmail.com>

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
package parser

import (
	"github.com/google/go-cmp/cmp"
	"github.com/tnagatomi/gh-mrlabel/option"
	"testing"
)

func TestLabel(t *testing.T) {
	type args struct {
		input string
	}
	tests := []struct {
		name    string
		args    args
		want    []option.Label
		wantErr bool
	}{
		{
			name: "Single label",
			args: args{
				input: "bug:ff0000:This is a bug",
			},
			want: []option.Label{
				{
					Name:        "bug",
					Color:       "ff0000",
					Description: "This is a bug",
				},
			},
			wantErr: false,
		},
		{
			name: "Single label without description",
			args: args{
				input: "bug:ff0000",
			},
			want: []option.Label{
				{
					Name:        "bug",
					Color:       "ff0000",
					Description: "",
				},
			},
			wantErr: false,
		},
		{
			name: "Multi labels",
			args: args{
				input: "bug:ff0000:This is a bug,enhancement:00ff00:This is an enhancement,question:0000ff:This is a question",
			},
			want: []option.Label{
				{
					Name:        "bug",
					Color:       "ff0000",
					Description: "This is a bug",
				},
				{
					Name:        "enhancement",
					Color:       "00ff00",
					Description: "This is an enhancement",
				},
				{
					Name:        "question",
					Color:       "0000ff",
					Description: "This is a question",
				},
			},
			wantErr: false,
		},
		{
			name: "Invalid label format",
			args: args{
				input: "bug:ff0000:This is a bug:invalid,enhancement:00ff00:This is an enhancement",
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "Invalid color format",
			args: args{
				input: "bug:ff000:This is a bug,enhancement:00ff00:This is an enhancement",
			},
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Label(tt.args.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("Label() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if diff := cmp.Diff(got, tt.want); diff != "" {
				t.Errorf("Label() mismatch (-got +want):\n%s", diff)
			}
		})
	}
}
