package config

import (
	"fmt"
	"log"
	"os"
	"path"
	"path/filepath"
	"strings"
)

var config map[string]string
var configName string

func init() {
	exePath, err := os.Executable()
	if err != nil {
		log.Println(err)
		return
	}
	configName = ".config"

	value := os.Getenv("CONFIG")
	if value != "" {
		configName = value
	}

	dir := filepath.Dir(exePath)
	configPath := path.Join(dir, "/"+configName)
	if _, err = os.Stat(configPath); err != nil {
		dir, err = os.Getwd()
		if err != nil {
			log.Println(err)
			return
		}
		configPath, err = find(dir)
		if err != nil {
			log.Fatalln(err)
			return
		}
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		log.Fatalln(err)
		return
	}

	str := strings.ReplaceAll(string(data), "\r\n", "\n")
	row := strings.Split(str, "\n")
	config = make(map[string]string)
	for _, s := range row {
		if s == "" || strings.HasPrefix(s, "#") {
			continue
		}
		cell := strings.Split(s, " = ")
		if len(cell) != 2 {
			err = fmt.Errorf("config row error:%s", s)
			log.Println(err)
			return
		}
		key := strings.Trim(cell[0], " ")
		value := strings.Trim(cell[1], " ")
		config[key] = value
	}
}

func find(dir string) (configPath string, err error) {
	configPath = path.Join(dir + "/" + configName)
	if _, err = os.Stat(configPath); err != nil {
		dir = path.Join(dir, "..")
		if dir == "/" {
			err = fmt.Errorf("%s not found", configName)
			return
		}
		return find(dir)
	}
	return
}
