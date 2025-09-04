package config

import "testing"

func TestConfig(t *testing.T) {
	tests := []struct {
		name string
	}{
		{
			name: "测试读取配置文件",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got1 := Value("MACHINE ID")
			got2 := Value("DATASOURCE")
			got3 := Value("ARR")
			got4 := Value("KEY")
			t.Logf("Config() = \n%v\n%v\n%v\n%v", got1, got2, got3, got4)
		})
	}
}
