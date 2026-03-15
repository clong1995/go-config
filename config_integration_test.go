package config

import (
	"os"
	"path/filepath"
	"testing"
)

// TestLoadConfig_Integration_WithRealFormat 是一个集成测试，用于验证配置加载的完整流程。
// 它覆盖了从文件查找到内容解析，再到通过 `Value` 函数读取的整个过程，
// 确保了各种基本格式（键值对、单行/多行数组）都能被正确处理。
func TestLoadConfig_Integration_WithRealFormat(t *testing.T) {
	// 步骤 1: 创建一个临时的 .config 文件。
	// 使用 t.TempDir() 可以确保测试结束后临时文件被自动清理，
	// 这使得测试完全独立，不依赖于外部环境。
	tempDir := t.TempDir()
	configContent := `
# 设备 ID，键包含空格
MACHINE ID = 80

# 多行数组
DATASOURCE = [
    account,
    access
]

# 单行数组
ARR = [aa,bb,cc]

# 简单的键值对
KEY = 123abc
`
	configPath := filepath.Join(tempDir, ".config")
	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatalf("创建临时配置文件失败: %v", err)
	}

	// 步骤 2: 切换当前工作目录到临时目录。
	// 这是为了模拟真实场景，让 `findConfigPath` 能够成功找到我们创建的配置文件。
	originalWD, err := os.Getwd()
	if err != nil {
		t.Fatalf("获取当前工作目录失败: %v", err)
	}
	if err := os.Chdir(tempDir); err != nil {
		t.Fatalf("切换工作目录到临时目录失败: %v", err)
	}
	// 使用 defer 语句确保测试执行完毕后，能够安全地恢复原始工作目录。
	defer func() {
		if err := os.Chdir(originalWD); err != nil {
			t.Errorf("恢复原始工作目录失败: %v", err)
		}
	}()

	// 步骤 3: 手动调用 loadConfig()，模拟包初始化时的配置加载行为。
	if err := loadConfig(); err != nil {
		t.Fatalf("loadConfig() 执行失败: %v", err)
	}

	// 步骤 4: 定义一系列测试用例，并进行断言。
	// 验证从配置文件中读取的各项配置是否符合预期。
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

// TestLoadConfig_Integration_ComplexArrayFormat 用于专门测试对复杂格式多行数组的解析能力。
// 这个测试验证了解析器是否能正确处理各种边缘情况，例如：
// - 同一行内包含多个元素
// - 元素前后存在多余的空格
// - 行尾带有逗号
func TestLoadConfig_Integration_ComplexArrayFormat(t *testing.T) {
	tempDir := t.TempDir()
	configContent := `
# 包含复杂格式的多行数组
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

	// 期望的结果是所有元素被正确拼接，并用单个逗号分隔。
	wantValue := "item1,item2,item3,item4,item5"
	gotValue, ok := Value("COMPLEX_ARR")

	if !ok {
		t.Fatalf("期望找到键 'COMPLEX_ARR'，但未找到")
	}
	if gotValue != wantValue {
		t.Errorf("对于复杂数组，Value() 的值 = %q, 期望是 %q", gotValue, wantValue)
	}
}
