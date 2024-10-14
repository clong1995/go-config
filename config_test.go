package config

import "testing"

func TestConfig(t *testing.T) {
	type args struct {
		key string
	}
	tests := []struct {
		name string
		args args
	}{
		{
			name: "测试读取配置文件",
			args: args{
				key: "PORT",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Config(tt.args.key)
			t.Logf("Config() = %v", got)
		})
	}
}
