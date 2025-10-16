package config

import (
	"fmt"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/clong1995/go-ansi-color"
)

var config map[string]string

// var configName string
var prefix = "config"

func init() {

	configName := ".config"

	envConfig := os.Getenv("CONFIG")
	if envConfig != "" {
		configName = envConfig
	}

	exePath, err := os.Executable()
	if err != nil {
		pcolor.PrintFatal(prefix, err.Error())
		return
	}

	dir := filepath.Dir(exePath)
	configPath := path.Join(dir, configName)
	if _, err = os.Stat(configPath); err != nil { //程序目录不存在
		//运行命令所在的目录，不一定是源码目录
		if dir, err = os.Getwd(); err != nil {
			pcolor.PrintFatal(prefix, err.Error())
			return
		}
		configPath = path.Join(dir, configName)
		if _, err = os.Stat(configPath); err != nil {
			pcolor.PrintFatal(prefix, err.Error())
			return
		}
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		pcolor.PrintFatal(prefix, err.Error())
		return
	}

	str := strings.ReplaceAll(string(data), "\r\n", "\n")

	//==数组成一行
	str = strings.ReplaceAll(str, "[\n", "[")
	str = strings.ReplaceAll(str, "\n]", "]")
	str = strings.ReplaceAll(str, ",\n", ",")
	str = strings.ReplaceAll(str, "    ", "")
	//==

	row := strings.Split(str, "\n")
	config = make(map[string]string)

	for _, s := range row {
		if s == "" || strings.HasPrefix(s, "#") {
			continue
		}

		cell := strings.Split(s, " = ")
		if len(cell) != 2 {
			err = fmt.Errorf("config row error:%s", s)
			pcolor.PrintFatal(prefix, err.Error())
			return
		}
		key := strings.Trim(cell[0], " ")
		value := strings.Trim(cell[1], " ")

		if strings.HasPrefix(value, "[") && strings.HasSuffix(value, "]") {
			value = strings.Trim(value, "[]")
		}

		config[key] = value
	}
}
