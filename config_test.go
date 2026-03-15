package config

import (
	"testing"
)

// setupTestConfig 是一个辅助函数，用于为单元测试设置一个模拟的（mock）配置。
// 这样可以使测试独立于外部文件，保证测试的稳定性和速度。
func setupTestConfig() {
	configMutex.Lock()
	defer configMutex.Unlock()
	config = map[string]string{
		"MACHINE_ID": "test_machine",
		"DATASOURCE": "test_source",
		"ARR":        "item1,item2,item3",
		"EMPTY_KEY":  "",
	}
}

// TestValue 是对 Value 函数的单元测试。
// 它不读取任何文件，而是直接在内存中测试 Value 函数的逻辑是否正确。
func TestValue(t *testing.T) {
	setupTestConfig() // 初始化模拟配置

	tests := []struct {
		name      string // 测试用例的名称
		key       string // 要查找的键
		wantValue string // 期望得到的值
		wantOk    bool   // 期望的 'ok' 状态
	}{
		{
			name:      "存在的键",
			key:       "MACHINE_ID",
			wantValue: "test_machine",
			wantOk:    true,
		},
		{
			name:      "另一个存在的键",
			key:       "DATASOURCE",
			wantValue: "test_source",
			wantOk:    true,
		},
		{
			name:      "类数组格式的值",
			key:       "ARR",
			wantValue: "item1,item2,item3",
			wantOk:    true,
		},
		{
			name:      "值为空字符串的键",
			key:       "EMPTY_KEY",
			wantValue: "",
			wantOk:    true,
		},
		{
			name:      "不存在的键",
			key:       "NON_EXISTENT",
			wantValue: "",
			wantOk:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotValue, gotOk := Value(tt.key)

			if gotOk != tt.wantOk {
				t.Errorf("Value() ok = %v, want %v", gotOk, tt.wantOk)
			}
			if gotValue != tt.wantValue {
				t.Errorf("Value() gotValue = %v, want %v", gotValue, tt.wantValue)
			}
		})
	}
}

// TestGet 是对 Get 函数的单元测试。
/*func TestGet(t *testing.T) {
	setupTestConfig()

	tests := []struct {
		name      string
		key       string
		wantValue string
	}{
		{
			name:      "存在的键",
			key:       "MACHINE_ID",
			wantValue: "test_machine",
		},
		{
			name:      "不存在的键",
			key:       "NON_EXISTENT",
			wantValue: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if gotValue := Get(tt.key); gotValue != tt.wantValue {
				t.Errorf("Get() = %v, want %v", gotValue, tt.wantValue)
			}
		})
	}
}*/
