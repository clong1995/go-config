// 防止循环依赖

package config

import (
	"fmt"
	"log"
	"os"
)

var shortLogger *log.Logger

func init() {
	prefix := "\033[1m[config]\033[0m "
	shortLogger = log.New(os.Stdout, prefix, 0)
}

func printFatal(err error) {
	shortLogger.Fatalln(fmt.Sprintf("\x1b[%dm%s\x1b[0m", 91, err.Error()))
}
