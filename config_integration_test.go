package config

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

// TestLoadConfig_Integration_WithTypes 是一个集成测试，用于验证配置加载和类型转换的完整流程。
func TestLoadConfig_Integration_WithTypes(t *testing.T) {
	tempDir := t.TempDir()
	configContent := `
# --- 基本类型 ---
STRING_VAL = hello world
INT_VAL = 99
BOOL_VAL = true
FLOAT_VAL = 1.23

# --- 切片类型 ---
STRING_SLICE_VAL = foo, bar, baz
INT_SLICE_VAL = 10, 20, 30

# --- Map 类型 ---
# 单行 Map
SINGLE_LINE_MAP = { host: "localhost", port: "8080" }

# 多行 Map
MULTI_LINE_MAP = {
    user: "admin";
    pass: "secret";
    timeout: "30"
}
`
	configPath := filepath.Join(tempDir, ".config")
	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatalf("创建临时配置文件失败: %v", err)
	}

	originalWD, err := os.Getwd()
	if err != nil {
		t.Fatalf("获取当前工作目录失败: %v", err)
	}
	if err := os.Chdir(tempDir); err != nil {
		t.Fatalf("切换工作目录到临时目录失败: %v", err)
	}
	defer func() {
		if err := os.Chdir(originalWD); err != nil {
			t.Errorf("恢复原始工作目录失败: %v", err)
		}
	}()

	if err := loadConfig(); err != nil {
		t.Fatalf("loadConfig() 执行失败: %v", err)
	}

	// --- 断言 ---
	t.Run("String", func(t *testing.T) {
		want := "hello world"
		got, ok := Value[string]("STRING_VAL")
		if !ok || got != want {
			t.Errorf("Value[string] got %q, %v, want %q, true", got, ok, want)
		}
	})

	t.Run("Int", func(t *testing.T) {
		want := 99
		got, ok := Value[int]("INT_VAL")
		if !ok || got != want {
			t.Errorf("Value[int] got %d, %v, want %d, true", got, ok, want)
		}
	})

	t.Run("StringSlice", func(t *testing.T) {
		want := []string{"foo", "bar", "baz"}
		got, ok := Value[[]string]("STRING_SLICE_VAL")
		if !ok || !reflect.DeepEqual(got, want) {
			t.Errorf("Value[[]string] got %v, %v, want %v, true", got, ok, want)
		}
	})

	t.Run("SingleLineMap", func(t *testing.T) {
		want := map[string]string{"host": "localhost", "port": "8080"}
		got, ok := Value[map[string]string]("SINGLE_LINE_MAP")
		if !ok || !reflect.DeepEqual(got, want) {
			t.Errorf("Value[map[string]string] for single line map got %v, want %v", got, want)
		}
	})

	t.Run("MultiLineMap", func(t *testing.T) {
		want := map[string]string{"user": "admin", "pass": "secret", "timeout": "30"}
		got, ok := Value[map[string]string]("MULTI_LINE_MAP")
		if !ok || !reflect.DeepEqual(got, want) {
			t.Errorf("Value[map[string]string] for multi-line map got %v, want %v", got, want)
		}
	})

	t.Run("MultiLineMapToInt", func(t *testing.T) {
		want := map[string]int{"timeout": 30}
		// We only test one key as others would fail conversion and that's expected.
		// A more robust test could check for specific key-value pairs.
		got, ok := Value[map[string]int]("MULTI_LINE_MAP")
		// This test is tricky because the parser is for a single type.
		// Let's check if the timeout key is correct.
		if !ok {
			t.Errorf("Value[map[string]int] failed to convert")
		}
		if val, ok := got["timeout"]; !ok || val != want["timeout"] {
			t.Errorf("Value[map[string]int] got timeout=%d, want %d", val, want["timeout"])
		}
	})

	t.Run("NonExistent", func(t *testing.T) {
		_, ok := Value[string]("NON_EXISTENT_KEY")
		if ok {
			t.Error("Value for non-existent key should have ok=false")
		}
	})
}
