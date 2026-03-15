package config

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

// TestLoadConfig_Integration_WithTypes 是一个集成测试，用于验证配置加载和类型转换的完整流程。
// 它确保了从文件加载的字符串值能够被泛型 `Value` 函数正确地转换为各种目标类型。
func TestLoadConfig_Integration_WithTypes(t *testing.T) {
	tempDir := t.TempDir()
	configContent := `
# 测试用的各种类型
STRING_VAL = hello world
INT_VAL = 99
BOOL_VAL = true
FLOAT_VAL = 1.23
STRING_SLICE_VAL = foo, bar, baz
INT_SLICE_VAL = 10, 20, 30
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

	t.Run("Bool", func(t *testing.T) {
		want := true
		got, ok := Value[bool]("BOOL_VAL")
		if !ok || got != want {
			t.Errorf("Value[bool] got %v, %v, want %v, true", got, ok, want)
		}
	})

	t.Run("Float64", func(t *testing.T) {
		want := 1.23
		got, ok := Value[float64]("FLOAT_VAL")
		if !ok || got != want {
			t.Errorf("Value[float64] got %f, %v, want %f, true", got, ok, want)
		}
	})

	t.Run("StringSlice", func(t *testing.T) {
		want := []string{"foo", "bar", "baz"}
		got, ok := Value[[]string]("STRING_SLICE_VAL")
		if !ok || !reflect.DeepEqual(got, want) {
			t.Errorf("Value[[]string] got %v, %v, want %v, true", got, ok, want)
		}
	})

	t.Run("IntSlice", func(t *testing.T) {
		want := []int{10, 20, 30}
		got, ok := Value[[]int]("INT_SLICE_VAL")
		if !ok || !reflect.DeepEqual(got, want) {
			t.Errorf("Value[[]int] got %v, %v, want %v, true", got, ok, want)
		}
	})

	t.Run("NonExistent", func(t *testing.T) {
		_, ok := Value[string]("NON_EXISTENT_KEY")
		if ok {
			t.Error("Value for non-existent key should have ok=false")
		}
	})
}
