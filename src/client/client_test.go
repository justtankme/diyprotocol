package main

import (
	"reflect"
	"testing"
)

func Test_getSendData(t *testing.T) {
	type args struct {
		clientNum int
		cmdNum    int
	}
	tests := []struct {
		name     string
		args     args
		wantData [][]string
	}{
		{
			name: "test1",
			args:     args{cmdNum: 1, clientNum: 1},
			wantData: [][]string{{""}},
		},
		{
			name: "test2",
			args:     args{cmdNum: 2, clientNum: 2},
			wantData: [][]string{{""}},
		},
		{
			name: "test3",
			args:     args{cmdNum: 3, clientNum: 3},
			wantData: [][]string{{""}},
		},
		{
			name: "test3",
			args:     args{cmdNum: 3, clientNum: 3},
			wantData: [][]string{{""}},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if gotData := getSendData(tt.args.clientNum, tt.args.cmdNum); !reflect.DeepEqual(gotData, tt.wantData) {
				t.Errorf("getSendData() = %v, want %v", gotData, tt.wantData)
			}
		})
	}
}
