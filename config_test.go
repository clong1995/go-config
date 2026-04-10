package config

import (
	"testing"
)

func TestValue(t *testing.T) {
	type args struct {
		key string
	}
	type testCase[T map[string]string] struct {
		name  string
		args  args
		want  T
		want1 bool
	}
	tests := []testCase[map[string]string]{
		{
			name: "Test map config",
			args: args{key: "SERVER"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, got1 := Value[map[string]string](tt.args.key)
			t.Logf("Value() got = %v", got)
			t.Logf("Value() got[user] = %v", got["user"])
			t.Logf("Value() got[account] = %v", got["account"])
			t.Logf("Value() got1 = %v", got1)
		})
	}
}
