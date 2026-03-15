package config

import (
	"reflect"
	"testing"
)

// setupTestConfig 是一个辅助函数，用于为单元测试设置一个模拟的（mock）配置。
// 这样可以使测试独立于外部文件，保证测试的稳定性和速度。
func setupTestConfig() {
	configMutex.Lock()
	defer configMutex.Unlock()
	config = map[string]string{
		"STRING_KEY":  "hello world",
		"INT_KEY":     "123",
		"BOOL_KEY":    "true",
		"FLOAT_KEY":   "3.14",
		"ARR_STRING":  "a,b,c",
		"ARR_INT":     "1, 2, 3",
		"EMPTY_KEY":   "",
		"INVALID_INT": "not-a-number",
	}
}

// TestValue_Generic_Types 是对泛型 Value 函数的单元测试。
// 它不读取任何文件，而是直接在内存中测试 Value 函数的类型转换逻辑是否正确。
func TestValue_Generic_Types(t *testing.T) {
	setupTestConfig() // 初始化模拟配置

	t.Run("String", func(t *testing.T) {
		got, ok := Value[string]("STRING_KEY")
		if !ok || got != "hello world" {
			t.Errorf("Value[string] got %q, %v, want %q, true", got, ok, "hello world")
		}
	})

	t.Run("Int", func(t *testing.T) {
		got, ok := Value[int]("INT_KEY")
		if !ok || got != 123 {
			t.Errorf("Value[int] got %d, %v, want %d, true", got, ok, 123)
		}
	})

	t.Run("Bool", func(t *testing.T) {
		got, ok := Value[bool]("BOOL_KEY")
		if !ok || got != true {
			t.Errorf("Value[bool] got %v, %v, want %v, true", got, ok, true)
		}
	})

	t.Run("Float64", func(t *testing.T) {
		got, ok := Value[float64]("FLOAT_KEY")
		if !ok || got != 3.14 {
			t.Errorf("Value[float64] got %f, %v, want %f, true", got, ok, 3.14)
		}
	})

	t.Run("StringSlice", func(t *testing.T) {
		want := []string{"a", "b", "c"}
		got, ok := Value[[]string]("ARR_STRING")
		if !ok || !reflect.DeepEqual(got, want) {
			t.Errorf("Value[[]string] got %v, %v, want %v, true", got, ok, want)
		}
	})

	t.Run("IntSlice", func(t *testing.T) {
		want := []int{1, 2, 3}
		got, ok := Value[[]int]("ARR_INT")
		if !ok || !reflect.DeepEqual(got, want) {
			t.Errorf("Value[[]int] got %v, %v, want %v, true", got, ok, want)
		}
	})

	t.Run("NonExistentKey", func(t *testing.T) {
		_, ok := Value[string]("NON_EXISTENT")
		if ok {
			t.Error("Value[string] for non-existent key should return ok=false")
		}
	})

	t.Run("InvalidConversion", func(t *testing.T) {
		_, ok := Value[int]("INVALID_INT")
		if ok {
			t.Error("Value[int] for invalid integer string should return ok=false")
		}
	})

	t.Run("EmptyValueForSlice", func(t *testing.T) {
		got, ok := Value[[]string]("EMPTY_KEY")
		if !ok || len(got) != 0 {
			t.Errorf("Value[[]string] for empty key should return empty slice, got %v", got)
		}
	})
}
