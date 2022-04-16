package recurringIssues

import (
	types "gitlab-issue-automation/types"
	"reflect"
	"testing"
)

func Test_parseMetadata(t *testing.T) {
	type args struct {
		contents []byte
	}
	tests := []struct {
		name    string
		args    args
		want    *types.Metadata
		wantErr bool
	}{
		{
			name: "Parses title",
			args: args{contents: ([]byte)(`---
title: Test Title
---
`)},
			want: &types.Metadata{
				Title: "Test Title",
			},
		},
		{
			name: "Parses description",
			args: args{contents: ([]byte)(`---
test: test
---
Test Description`)},
			want: &types.Metadata{
				Description: "Test Description",
			},
		},
		{
			name: "Parses confidential",
			args: args{contents: ([]byte)(`---
confidential: true
---
`)},
			want: &types.Metadata{
				Confidential: true,
			},
		},
		{
			name: "Parses assignee",
			args: args{contents: ([]byte)(`---
assignees: [ "assignee1" ]
---
`)},
			want: &types.Metadata{
				Assignees: []string{"assignee1"},
			},
		},
		{
			name: "Parses assignees",
			args: args{contents: ([]byte)(`---
assignees: [ "assignee1", "assignee2" ]
---
`)},
			want: &types.Metadata{
				Assignees: []string{"assignee1", "assignee2"},
			},
		},
		{
			name: "Parses label",
			args: args{contents: ([]byte)(`---
labels: [ "label1" ]
---
`)},
			want: &types.Metadata{
				Labels: []string{"label1"},
			},
		},
		{
			name: "Parses labels",
			args: args{contents: ([]byte)(`---
labels: [ "label1", "label2" ]
---
`)},
			want: &types.Metadata{
				Labels: []string{"label1", "label2"},
			},
		},
		{
			name: "Parses dueindays",
			args: args{contents: ([]byte)(`---
duein: 24h
---
`)},
			want: &types.Metadata{
				DueIn: "24h",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseMetadata(tt.args.contents)
			if (err != nil) != tt.wantErr {
				t.Errorf("parseMetadata() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("parseMetadata() = %v, want %v", got, tt.want)
			}
		})
	}
}
