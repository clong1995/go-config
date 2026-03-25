package config

import (
	"bufio"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/pkg/errors"

	"github.com/clong1995/go-ansi-color"
)

var prefix = "config"

// init 函数在包被导入时自动执行。
// 它调用 loadConfig 来加载配置，如果加载过程中发生任何错误，
// 程序将打印致命错误信息并立即终止。
func init() {
	if err := loadConfig(); err != nil {
		pcolor.PrintFatal(prefix, "%+v", err)
	}
}

// loadConfig 是配置加载的核心函数。它负责查找、读取、并解析配置文件。
//
// 解析逻辑支持以下格式：
//   - 键值对 (例如: KEY = VALUE)
//   - 单行/多行数组 (例如: ARR = [item1, item2])
//   - 单行/多行 Map (例如: MAP = {key1:val1; key2:val2})
//
// 函数会忽略空行和以 '#' 开头的注释行。
func loadConfig() error {
	// 默认的配置文件名是 .config，但可以通过设置 `CONFIG` 环境变量来覆盖。
	configName := ".config"
	if envConfig := os.Getenv("CONFIG"); envConfig != "" {
		configName = envConfig
	}

	// 查找配置文件的绝对路径。
	configPath, err := findConfigPath(configName)
	if err != nil {
		return errors.Wrap(err, "查找配置文件路径失败")
	}

	file, err := os.Open(configPath)
	if err != nil {
		return errors.Wrap(err, "打开配置文件失败")
	}
	defer func(f *os.File) {
		_ = f.Close()
	}(file)

	newConfig := make(map[string]string)
	scanner := bufio.NewScanner(file)

	// 用于处理多行块（数组或 map）的状态变量。
	var multiLineKey string
	var multiLineContent string
	var inMultiLineBlock bool
	var blockEndChar string // "]" 代表数组, "}" 代表 map

	for scanner.Scan() {
		line := scanner.Text()
		trimmedLine := strings.TrimSpace(line)

		// 如果处于多行块的解析状态中。
		if inMultiLineBlock {
			if trimmedLine == blockEndChar {
				// 遇到块结束符，处理并存储内容。
				var finalContent string
				if blockEndChar == "]" { // 数组块
					// 对于数组，将内容用逗号连接。
					var cleanedElements []string
					elements := strings.Split(multiLineContent, ",")
					for _, el := range elements {
						trimmedEl := strings.TrimSpace(el)
						if trimmedEl != "" {
							cleanedElements = append(cleanedElements, trimmedEl)
						}
					}
					finalContent = strings.Join(cleanedElements, ",")
				} else { // Map 块
					// 对于 map，直接使用拼接好的内容。
					finalContent = multiLineContent
				}
				newConfig[multiLineKey] = finalContent
				// 重置状态。
				inMultiLineBlock = false
				multiLineKey = ""
				multiLineContent = ""
				blockEndChar = ""
			} else if trimmedLine != "" && !strings.HasPrefix(trimmedLine, "#") {
				// 将非空、非注释的行内容拼接到 multiLineContent 中。
				if blockEndChar == "}" {
					// 对 map 使用分号作为条目分隔符，以增加健壮性。
					multiLineContent += trimmedLine + ";"
				} else {
					// 对于数组追加逗号拼接，防止用户未在行尾加逗号导致元素粘连。
					multiLineContent += trimmedLine + ","
				}
			}
			continue
		}

		// 忽略空行和注释行。
		if trimmedLine == "" || strings.HasPrefix(trimmedLine, "#") {
			continue
		}

		// 解析 "key = value" 格式的键值对。
		parts := strings.SplitN(trimmedLine, "=", 2)
		if len(parts) != 2 {
			continue
		}

		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])

		// 检查值的格式，以确定如何处理。
		if value == "[" {
			inMultiLineBlock = true
			multiLineKey = key
			blockEndChar = "]"
		} else if value == "{" {
			inMultiLineBlock = true
			multiLineKey = key
			blockEndChar = "}"
		} else if (strings.HasPrefix(value, "[") && strings.HasSuffix(value, "]")) ||
			(strings.HasPrefix(value, "{") && strings.HasSuffix(value, "}")) {
			// 对于单行数组或 map，直接剥离两端的括号，避免使用 strings.Trim 误删内部字符。
			newConfig[key] = value[1 : len(value)-1]
		} else {
			// 普通的键值对。
			newConfig[key] = value
		}
	}

	if err = scanner.Err(); err != nil {
		return errors.Wrap(err, "读取文件时发生错误")
	}

	// 检查在文件末尾是否存在未闭合的块。
	if inMultiLineBlock {
		return errors.Errorf("配置文件在结尾处，键 '%s' 的块没有闭合", multiLineKey)
	}

	// 在修改全局配置前加锁，以确保线程安全。
	configMutex.Lock()
	// 将解析出的新配置赋值给全局变量。
	config = newConfig
	configMutex.Unlock()

	// 清空所有旧的类型转换缓存，因为配置已更新。
	cache.Clear()
	return nil
}

// findConfigPath 智能地查找配置文件的路径。
// 查找顺序：1. 程序可执行文件目录。2. 当前工作目录及所有上层目录。
func findConfigPath(configName string) (string, error) {
	// 1. 检查可执行文件所在的目录。
	if exePath, err := os.Executable(); err == nil {
		configPath := path.Join(filepath.Dir(exePath), configName)
		if _, err = os.Stat(configPath); err == nil {
			return configPath, nil
		}
	}

	// 2. 从当前工作目录开始，向上递归查找。
	if wd, err := os.Getwd(); err == nil {
		currentDir := wd
		for {
			configPath := path.Join(currentDir, configName)
			if _, err = os.Stat(configPath); err == nil {
				return configPath, nil
			}
			parentDir := filepath.Dir(currentDir)
			if parentDir == currentDir { // 到达根目录，停止查找。
				break
			}
			currentDir = parentDir
		}
	}

	return "", errors.Errorf("在任何预定位置都找不到配置文件 '%s'", configName)
}
