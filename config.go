package config

import "sync"

var (
	// configMutex 是一个读写互斥锁，用于在并发环境中保护全局配置 `config` 的线程安全。
	configMutex sync.RWMutex
	// config 是一个全局映射，以键值对的形式存储所有配置项。
	// 这个变量由 `loadConfig` 函数在包初始化时填充。
	config map[string]string
)

// Value 根据键名安全地获取配置值。
// 这是一个线程安全的函数，因为它内部使用了读锁。
//
// 返回值:
//   - string: 键对应的值。如果键不存在，则返回空字符串。
//   - bool:   一个布尔值，如果键存在，则为 true，否则为 false。
func Value(key string) (string, bool) {
	configMutex.RLock() // 加读锁，允许多个并发的读操作
	defer configMutex.RUnlock()
	val, ok := config[key]
	return val, ok
}

// Get 是 Value 的一个简化版本，仅返回配置值。
// 如果键不存在，它会返回一个空字符串。
//
// Deprecated: 为了更明确地处理“键不存在”和“值为空字符串”这两种不同情况，
// 推荐使用 Value 函数。此函数无法区分这两种场景。
/*func Get(key string) string {
	val, _ := Value(key)
	return val
}*/
