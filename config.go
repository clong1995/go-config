package config

import "sync"

var (
	// configMutex 是一个读写互斥锁，用于保护对 config map 的并发访问。
	configMutex sync.RWMutex
	// config 是一个全局的 map，用于存储从配置文件中读取的键值对。
	config map[string]string
)

// Value 根据键名获取配置值。
// 它返回两个值：
// 1. string: 键对应的值。如果键不存在，则返回空字符串。
// 2. bool: 一个布尔值，表示键是否存在。
// 这是一个线程安全的函数。
func Value(key string) (string, bool) {
	configMutex.RLock() // 加读锁，允许多个并发读取
	defer configMutex.RUnlock()
	val, ok := config[key]
	return val, ok
}

// Get 是 Value 的一个简化版本，仅返回配置值。
// 如果键不存在，它会返回一个空字符串。
//
// Deprecated: 为了更安全地检查键是否存在，推荐使用 Value 函数。
// Get 函数无法区分“值为空字符串”和“键不存在”这两种情况。
func Get(key string) string {
	val, _ := Value(key)
	return val
}
