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

func init() {
	// 在包初始化时加载配置，如果失败则直接终止程序
	if err := loadConfig(); err != nil {
		pcolor.PrintFatal(prefix, "%+v", err)
	}
}

// loadConfig 负责查找、读取并解析配置文件。
// 它支持简单的键值对、单行数组和多行数组格式。
func loadConfig() error {
	// 默认配置文件名为 .config，但可以通过 CONFIG 环境变量覆盖
	configName := ".config"
	if envConfig := os.Getenv("CONFIG"); envConfig != "" {
		configName = envConfig
	}

	// 查找配置文件的最终路径
	configPath, err := findConfigPath(configName)
	if err != nil {
		return errors.Wrap(err, "find config path")
	}

	file, err := os.Open(configPath)
	if err != nil {
		return errors.Wrap(err, "failed to open config file")
	}
	defer func(f *os.File) {
		_ = f.Close()
	}(file)

	// 加锁以保证并发安全
	configMutex.Lock()
	defer configMutex.Unlock()

	newConfig := make(map[string]string)
	scanner := bufio.NewScanner(file)

	// 用于处理多行数组的状态变量
	var arrayKey string       // 当前正在解析的多行数组的键
	var arrayContent []string // 临时存储多行数组的内容

	for scanner.Scan() {
		line := scanner.Text()
		trimmedLine := strings.TrimSpace(line)

		// 如果 arrayKey 不为空，说明我们正处于一个多行数组的解析过程中
		if arrayKey != "" {
			if trimmedLine == "]" {
				// 数组结束
				newConfig[arrayKey] = strings.Join(arrayContent, ",")
				arrayKey = ""
				arrayContent = nil
			} else if trimmedLine != "" && !strings.HasPrefix(trimmedLine, "#") {
				// 数组中的一个元素
				element := strings.TrimSuffix(trimmedLine, ",")
				arrayContent = append(arrayContent, strings.TrimSpace(element))
			}
			continue // 继续处理下一行
		}

		// 忽略空行和注释行
		if trimmedLine == "" || strings.HasPrefix(trimmedLine, "#") {
			continue
		}

		// 解析键值对
		parts := strings.SplitN(trimmedLine, "=", 2)
		if len(parts) != 2 {
			continue // 格式错误的行，直接跳过
		}

		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])

		if value == "[" {
			// 侦测到多行数组的开始
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
			// 解析简单的键值对
			newConfig[key] = value
		}
	}

	if err = scanner.Err(); err != nil {
		return errors.Wrap(err, "scanner error")
	}

	// 检查是否有未闭合的多行数组
	if arrayKey != "" {
		return errors.Errorf("配置文件结尾处，键 '%s' 的数组没有闭合", arrayKey)
	}

	// 将解析出的配置赋值给全局变量
	config = newConfig
	return nil
}

// findConfigPath 查找配置文件的路径。
// 查找顺序:
// 1. 程序可执行文件所在的目录。
// 2. 当前工作目录，并向上递归查找直到根目录。
func findConfigPath(configName string) (string, error) {
	// 1. 检查可执行文件目录
	if exePath, err := os.Executable(); err == nil {
		exeDir := filepath.Dir(exePath)
		configPath := path.Join(exeDir, configName)
		if _, err = os.Stat(configPath); err == nil {
			return configPath, nil
		}
	}

	// 2. 检查当前工作目录及所有上层目录
	if wd, err := os.Getwd(); err == nil {
		currentDir := wd
		for {
			configPath := path.Join(currentDir, configName)
			if _, err = os.Stat(configPath); err == nil {
				return configPath, nil
			}

			parentDir := filepath.Dir(currentDir)
			if parentDir == currentDir { // 到达根目录, 停止查找
				break
			}
			currentDir = parentDir
		}
	}

	return "", errors.Errorf("找不到配置文件 '%s'", configName)
}
