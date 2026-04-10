package config

import (
	"reflect"
	"strconv"
	"strings"
	"sync"
)

var (
	// configMutex 是一个读写互斥锁，用于保护全局配置 `config` 的线程安全。
	configMutex sync.RWMutex
	// config 是一个全局映射，以键值对的形式存储所有从文件中解析出的原始字符串配置。
	config map[string]string
	// cache 是一个线程安全的缓存，用于存储类型转换的结果，避免重复进行昂贵的转换操作。
	cache = valueCache{data: make(map[string]any)}
)

// valueCache 是一个为类型转换结果设计的线程安全缓存。
type valueCache struct {
	mutex sync.RWMutex
	data  map[string]any // 缓存的 key 格式为 "configKey:typeName"，例如 "SERVER:map[string]string"
}

// GetOrSet 是获取或设置缓存的核心方法。
// 它采用“双重检查锁定”模式来确保高效率和线程安全。
// 如果缓存命中，它直接返回缓存的值；如果未命中，则执行 factory 函数来生成值，并将结果存入缓存后再返回。
func (c *valueCache) getOrSet(key string, factory func() (any, bool)) (any, bool) {
	// 快速路径：使用读锁检查缓存是否存在，允许多个 goroutine 并发读取。
	c.mutex.RLock()
	val, exists := c.data[key]
	c.mutex.RUnlock()
	if exists {
		return val, true
	}

	// 慢速路径：使用写锁来执行转换和写入操作。
	c.mutex.Lock()
	defer c.mutex.Unlock()

	// 双重检查：在等待写锁期间，可能已有其他 goroutine 完成了操作，因此需要再次检查。
	val, exists = c.data[key]
	if exists {
		return val, true
	}

	// 执行昂贵的转换操作（由 factory 函数提供）。
	val, ok := factory()
	if ok {
		// 只有在转换成功时才缓存结果。
		c.data[key] = val
	}
	return val, ok
}

// Clear 重置缓存。此操作在配置重新加载时被调用，以确保缓存与配置同步。
func (c *valueCache) Clear() {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	c.data = make(map[string]any)
}

// Value 根据键名安全地获取配置，并将其转换为调用者指定的类型 T。
// 这个函数是库的干净入口点，它将复杂的缓存和转换逻辑委托给其他部分处理。
func Value[T any](key string) (T, bool) {
	var zero T
	// 使用反射获取类型名称，以创建唯一的缓存键，例如 "PORT:int"。
	cacheKey := key + ":" + reflect.TypeFor[T]().String()

	// 调用缓存的 getOrSet 方法。真正的转换逻辑（doConversion）只在缓存未命中时执行。
	val, ok := cache.getOrSet(cacheKey, func() (any, bool) {
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
	// 利用类型开关，根据调用者期望的类型 T 执行相应的转换逻辑。
	switch any(zero).(type) {
	// --- 基础类型 ---
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

	// --- 切片类型 ---
	case []string:
		return parseSlice(valStr, func(s string) (string, error) { return s, nil })
	case []int:
		return parseSlice(valStr, strconv.Atoi)
	case []int64:
		return parseSlice(valStr, func(s string) (int64, error) { return strconv.ParseInt(s, 10, 64) })
	case []float64:
		return parseSlice(valStr, func(s string) (float64, error) { return strconv.ParseFloat(s, 64) })
	case []bool:
		return parseSlice(valStr, strconv.ParseBool)

	// --- Map 类型 ---
	case map[string]string:
		return parseMap(valStr, func(s string) (string, error) { return s, nil })
	case map[string]int:
		return parseMap(valStr, strconv.Atoi)
	case map[string]int64:
		return parseMap(valStr, func(s string) (int64, error) { return strconv.ParseInt(s, 10, 64) })
	case map[string]float64:
		return parseMap(valStr, func(s string) (float64, error) { return strconv.ParseFloat(s, 64) })
	case map[string]bool:
		return parseMap(valStr, strconv.ParseBool)

	default:
		return nil, false
	}
}

// parseSlice 是一个泛型辅助函数，用于将逗号分隔的字符串解析为指定类型的切片。
// 它接受一个字符串和一个 `parser` 函数，该函数负责将单个字符串片段转换为目标类型 T。
func parseSlice[T any](valStr string, parser func(string) (T, error)) (any, bool) {
	if valStr == "" {
		return []T{}, true
	}
	parts := strings.Split(valStr, ",")
	result := make([]T, 0, len(parts))
	for _, p := range parts {
		trimmed := strings.TrimSpace(p)
		if trimmed == "" {
			continue // 忽略空元素，提升对尾部逗号等格式的容错性。
		}
		val, err := parser(trimmed)
		if err != nil {
			return nil, false // 如果任何片段解析失败，则整个切片解析失败。
		}
		result = append(result, val)
	}
	return result, true
}

// parseMap 是一个泛型辅助函数，用于将特定格式的字符串解析为指定类型的 map。
// 字符串格式为 "key1:val1,key2:val2"。
// 它接受一个字符串和一个 `parser` 函数，该函数负责将值的字符串部分转换为目标类型 T。
func parseMap[T any](valStr string, parser func(string) (T, error)) (any, bool) {
	if valStr == "" {
		return map[string]T{}, true
	}
	result := make(map[string]T)
	// 使用逗号作为 map 条目之间的分隔符，与切片解析保持一致。
	entries := strings.Split(valStr, ",")
	for _, entry := range entries {
		entry = strings.TrimSpace(entry)
		if entry == "" {
			continue
		}
		// 使用冒号作为键和值之间的分隔符。
		parts := strings.SplitN(entry, ":", 2)
		if len(parts) != 2 {
			return nil, false // 格式错误的条目。
		}
		key := strings.TrimSpace(parts[0])
		val, err := parser(strings.TrimSpace(parts[1]))
		if err != nil {
			return nil, false // 如果任何值的解析失败，则整个 map 解析失败。
		}
		result[key] = val
	}
	return result, true
}
