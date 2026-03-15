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
// 这种设计适用于那些强依赖配置才能启动的应用程序。
func init() {
	if err := loadConfig(); err != nil {
		pcolor.PrintFatal(prefix, "%+v", err)
	}
}

// loadConfig 是配置加载的核心函数。它负责查找、读取、解析配置文件，
// 并将结果填充到全局的 `config` 变量中。
//
// 解析逻辑支持以下格式：
//   - 键值对 (例如: KEY = VALUE)
//   - 单行数组 (例如: ARR = [item1, item2])
//   - 多行数组 (例如: ARR = [ item1, item2 ])
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

	// 在修改全局配置前加锁，以确保线程安全。
	configMutex.Lock()
	defer configMutex.Unlock()

	newConfig := make(map[string]string)
	scanner := bufio.NewScanner(file)

	// 用于处理多行数组解析的状态变量。
	var arrayKey string     // 存储当前正在解析的多行数组的键。
	var arrayContent string // 将多行数组的所有行拼接成一个长字符串，以便后续处理。

	for scanner.Scan() {
		line := scanner.Text()
		trimmedLine := strings.TrimSpace(line)

		// 如果 arrayKey 不为空，表示当前正处于一个多行数组的解析状态。
		if arrayKey != "" {
			if trimmedLine == "]" {
				// 遇到数组结束符 ']'，开始处理拼接好的数组内容。
				var cleanedElements []string
				elements := strings.Split(arrayContent, ",")
				for _, el := range elements {
					trimmedEl := strings.TrimSpace(el)
					if trimmedEl != "" { // 过滤掉因多余逗号产生的空元素。
						cleanedElements = append(cleanedElements, trimmedEl)
					}
				}
				newConfig[arrayKey] = strings.Join(cleanedElements, ",")
				arrayKey = ""     // 重置状态，退出数组解析模式。
				arrayContent = "" // 清空内容。
			} else if trimmedLine != "" && !strings.HasPrefix(trimmedLine, "#") {
				// 将非空、非注释的行内容拼接到 arrayContent 中。
				arrayContent += trimmedLine
			}
			continue // 继续扫描下一行。
		}

		// 忽略空行和注释行。
		if trimmedLine == "" || strings.HasPrefix(trimmedLine, "#") {
			continue
		}

		// 解析 "key = value" 格式的键值对。
		parts := strings.SplitN(trimmedLine, "=", 2)
		if len(parts) != 2 {
			continue // 如果行不包含 '='，则视为格式错误并跳过。
		}

		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])

		if value == "[" {
			// 侦测到多行数组的开始。
			arrayKey = key
		} else if strings.HasPrefix(value, "[") && strings.HasSuffix(value, "]") {
			// 解析单行数组，例如: ARR = [aa,bb,cc]
			content := strings.Trim(value, "[]")
			elements := strings.Split(content, ",")
			var cleanedElements []string
			for _, el := range elements {
				cleanedElements = append(cleanedElements, strings.TrimSpace(el))
			}
			newConfig[key] = strings.Join(cleanedElements, ",")
		} else {
			// 解析简单的键值对。
			newConfig[key] = value
		}
	}

	if err = scanner.Err(); err != nil {
		return errors.Wrap(err, "读取文件时发生错误")
	}

	// 检查在文件末尾是否存在未闭合的多行数组。
	if arrayKey != "" {
		return errors.Errorf("配置文件在结尾处，键 '%s' 的数组没有闭合", arrayKey)
	}

	// 将解析出的新配置赋值给全局变量。
	config = newConfig
	return nil
}

// findConfigPath 智能地查找配置文件的路径。
//
// 查找顺序如下:
//  1. 首先，在程序可执行文件所在的目录中查找。
//  2. 如果未找到，则从当前工作目录开始，向上逐级递归查找，直至文件系统的根目录。
//
// 如果在所有指定位置都找不到配置文件，将返回一个错误。
func findConfigPath(configName string) (string, error) {
	// 1. 检查可执行文件所在的目录。
	if exePath, err := os.Executable(); err == nil {
		exeDir := filepath.Dir(exePath)
		configPath := path.Join(exeDir, configName)
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
			if parentDir == currentDir { // 如果到达根目录，则停止查找。
				break
			}
			currentDir = parentDir
		}
	}

	return "", errors.Errorf("在任何预定位置都找不到配置文件 '%s'", configName)
}
