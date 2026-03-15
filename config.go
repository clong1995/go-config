package config

import (
	"reflect"
	"strconv"
	"strings"
	"sync"
)

var (
	// configMutex 是一个读写互斥锁，用于在并发环境中保护全局配置 `config` 的线程安全。
	configMutex sync.RWMutex
	// config 是一个全局映射，以键值对的形式存储所有配置项。
	// 这个变量由 `loadConfig` 函数在包初始化时填充。
	config map[string]string
	// cache 是一个线程安全的缓存，用于存储类型转换的结果，避免重复计算。
	cache = valueCache{data: make(map[string]any)}
)

// valueCache 是一个为类型转换结果设计的线程安全缓存。
type valueCache struct {
	mutex sync.RWMutex
	data  map[string]any // 缓存的 key 格式为 "configKey:typeName"
}

// GetOrSet 是获取或设置缓存的核心方法。
// 它首先尝试从缓存中获取值，如果未命中，则执行 factory 函数来生成值，
// 并将结果存入缓存后返回。
func (c *valueCache) GetOrSet(key string, factory func() (any, bool)) (any, bool) {
	// 快速路径：使用读锁检查缓存是否存在。
	c.mutex.RLock()
	val, exists := c.data[key]
	c.mutex.RUnlock()
	if exists {
		return val, true
	}

	// 慢速路径：使用写锁来执行转换和写入操作。
	c.mutex.Lock()
	defer c.mutex.Unlock()

	// 双重检查锁定：在等待写锁期间，可能已有其他 goroutine 完成了操作。
	val, exists = c.data[key]
	if exists {
		return val, true
	}

	// 执行昂贵的转换操作。
	val, ok := factory()
	if ok {
		// 只有在转换成功时才缓存结果。
		c.data[key] = val
	}
	return val, ok
}

// Clear 重置缓存，通常在配置重新加载时调用。
func (c *valueCache) Clear() {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	c.data = make(map[string]any)
}

// Value 根据键名安全地获取配置，并将其转换为指定的类型 T。
// 这个函数是干净的入口点，它将复杂的缓存和转换逻辑委托给其他部分处理。
func Value[T any](key string) (T, bool) {
	var zero T
	// 使用反射获取类型名称，以创建唯一的缓存键，例如 "PORT:int"。
	cacheKey := key + ":" + reflect.TypeOf(zero).String()

	// 调用缓存的 GetOrSet 方法。转换逻辑（doConversion）只在缓存未命中时执行。
	val, ok := cache.GetOrSet(cacheKey, func() (any, bool) {
		return doConversion[T](key)
	})

	if !ok {
		return zero, false
	}
	// 从缓存的 `any` 类型安全地断言回目标类型 T。
	return val.(T), true
}

// doConversion 包含了从字符串到目标类型 T 的实际转换逻辑。
// 这个函数只应在缓存未命中时被调用。
func doConversion[T any](key string) (any, bool) {
	configMutex.RLock()
	valStr, ok := config[key]
	configMutex.RUnlock()

	if !ok {
		return nil, false
	}

	var zero T
	switch any(zero).(type) {
	case string:
		return valStr, true
	case int:
		i, err := strconv.Atoi(valStr)
		if err != nil {
			return nil, false
		}
		return i, true
	case int64:
		i64, err := strconv.ParseInt(valStr, 10, 64)
		if err != nil {
			return nil, false
		}
		return i64, true
	case float64:
		f64, err := strconv.ParseFloat(valStr, 64)
		if err != nil {
			return nil, false
		}
		return f64, true
	case bool:
		b, err := strconv.ParseBool(valStr)
		if err != nil {
			return nil, false
		}
		return b, true
	case []string:
		if valStr == "" {
			return []string{}, true
		}
		return strings.Split(valStr, ","), true
	case []int:
		parts := strings.Split(valStr, ",")
		result := make([]int, 0, len(parts))
		for _, p := range parts {
			i, err := strconv.Atoi(strings.TrimSpace(p))
			if err != nil {
				return nil, false
			}
			result = append(result, i)
		}
		return result, true
	case []int64:
		parts := strings.Split(valStr, ",")
		result := make([]int64, 0, len(parts))
		for _, p := range parts {
			i64, err := strconv.ParseInt(strings.TrimSpace(p), 10, 64)
			if err != nil {
				return nil, false
			}
			result = append(result, i64)
		}
		return result, true
	case []float64:
		parts := strings.Split(valStr, ",")
		result := make([]float64, 0, len(parts))
		for _, p := range parts {
			f64, err := strconv.ParseFloat(strings.TrimSpace(p), 64)
			if err != nil {
				return nil, false
			}
			result = append(result, f64)
		}
		return result, true
	case []bool:
		parts := strings.Split(valStr, ",")
		result := make([]bool, 0, len(parts))
		for _, p := range parts {
			b, err := strconv.ParseBool(strings.TrimSpace(p))
			if err != nil {
				return nil, false
			}
			result = append(result, b)
		}
		return result, true
	default:
		return nil, false
	}
}
