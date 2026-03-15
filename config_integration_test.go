package config

import (
	"os"
	"path/filepath"
	"testing"
)

// TestLoadConfig_Integration_WithRealFormat 是一个集成测试。
// 它的目的是验证从“查找文件”到“解析内容”再到“通过 Value 函数读取”的整个流程是否能正确工作。
// 这个测试模拟了真实的文件结构和内容格式。
func TestLoadConfig_Integration_WithRealFormat(t *testing.T) {
	// 步骤 1: 在一个临时目录中创建一个虚拟的 .config 文件。
	// 这样做可以避免测试依赖于项目中的真实配置文件，使测试更稳定。
	tempDir := t.TempDir()
	configContent := `
# 设备id
MACHINE ID = 80

# 多行数据源
DATASOURCE = [
    account,
    access
]

# 单行数组
ARR = [aa,bb,cc]

# 简单键值
KEY = 123abc
`
	configPath := filepath.Join(tempDir, ".config")
	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatalf("创建临时配置文件失败: %v", err)
	}

	// 步骤 2: 将当前工作目录切换到临时目录。
	// 这是为了让 findConfigPath 函数能够找到我们刚刚创建的临时文件。
	originalWD, err := os.Getwd()
	if err != nil {
		t.Fatalf("获取当前工作目录失败: %v", err)
	}
	if err := os.Chdir(tempDir); err != nil {
		t.Fatalf("切换工作目录失败: %v", err)
	}
	// 使用 defer 确保测试结束后能恢复原始的工作目录，避免影响其他测试。
	defer func() {
		if err := os.Chdir(originalWD); err != nil {
			t.Errorf("恢复工作目录失败: %v", err)
		}
	}()

	// 步骤 3: 手动触发配置加载流程，这相当于模拟了包初始化时的 init() 行为。
	if err := loadConfig(); err != nil {
		t.Fatalf("loadConfig() 执行失败: %v", err)
	}

	// 步骤 4: 断言（Assert），检查文件中的配置项是否被正确加载和解析。
	testCases := []struct {
		name      string
		key       string
		wantValue string
		wantOk    bool
	}{
		{"带空格的键", "MACHINE ID", "80", true},
		{"多行数组", "DATASOURCE", "account,access", true},
		{"单行数组", "ARR", "aa,bb,cc", true},
		{"简单的键值对", "KEY", "123abc", true},
		{"不存在的键", "NON_EXISTENT", "", false},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			gotValue, gotOk := Value(tc.key)
			if gotOk != tc.wantOk {
				t.Errorf("Value() 的 ok 返回值 = %v, 期望是 %v", gotOk, tc.wantOk)
			}
			if gotValue != tc.wantValue {
				t.Errorf("Value() 的值 = %q, 期望是 %q", gotValue, tc.wantValue)
			}
		})
	}
}

// TestLoadConfig_Integration_ComplexArrayFormat 用于测试更复杂的多行数组格式。
// 这个测试验证了解析器是否能正确处理各种边缘情况，例如行尾逗号、元素周围的空格以及同一行内的多个元素。
func TestLoadConfig_Integration_ComplexArrayFormat(t *testing.T) {
	tempDir := t.TempDir()
	configContent := `
# 复杂格式的多行数组
COMPLEX_ARR = [
    item1, item2,  # 同一行有多个元素
    item3  ,       # 元素后有空格和逗号
      item4,       # 元素前有空格
    item5          # 最后一个元素，没有逗号
]
`
	configPath := filepath.Join(tempDir, ".config")
	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatalf("创建临时配置文件失败: %v", err)
	}

	originalWD, _ := os.Getwd()
	_ = os.Chdir(tempDir)
	defer func() { _ = os.Chdir(originalWD) }()

	if err := loadConfig(); err != nil {
		t.Fatalf("loadConfig() 执行失败: %v", err)
	}

	wantValue := "item1,item2,item3,item4,item5"
	gotValue, ok := Value("COMPLEX_ARR")

	if !ok {
		t.Fatalf("期望找到键 'COMPLEX_ARR'，但未找到")
	}
	if gotValue != wantValue {
		t.Errorf("对于复杂数组，Value() 的值 = %q, 期望是 %q", gotValue, wantValue)
	}
}
