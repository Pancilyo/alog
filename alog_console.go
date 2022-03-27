package alog

import (
	"fmt"
	"time"
)

// consoleLogger 日志对象
type consoleLogger struct {
	level      Level
	timeFormat string
}

// newConsoleLogger 构造函数
func newConsoleLogger() *consoleLogger {
	return &consoleLogger{
		level:      defaultLevel,
		timeFormat: defaultTimeFormat,
	}
}

// enable 查看等级是否符合ConsoleLogger初始化的等级
func (c *consoleLogger) enable(logeLevel Level) bool {
	return c.level <= logeLevel
}

// log 格式化输出内容
func (c *consoleLogger) log(lv Level, format interface{}, a ...interface{}) {
	if c.enable(lv) {
		msg := fmt.Sprintf(fmt.Sprintf("%v", format), a...)
		now := time.Now()
		fileName, funcName, lineNo := traceInfo(4)
		fmt.Printf("[%s] [%s] [%s:%s:%d] %s\n", now.Format(c.timeFormat),
			l2S(lv), fileName, funcName, lineNo, msg)
	}
}
func (c *consoleLogger) Debug(format interface{}, a ...interface{}) {
	c.log(DEBUG, format, a...)
}
func (c *consoleLogger) Info(format interface{}, a ...interface{}) {
	c.log(INFO, format, a...)
}
func (c *consoleLogger) Warn(format interface{}, a ...interface{}) {
	c.log(WARN, format, a...)
}
func (c *consoleLogger) Error(format interface{}, a ...interface{}) {
	c.log(ERROR, format, a...)
}
func (c *consoleLogger) Fatal(format interface{}, a ...interface{}) {
	c.log(FATAL, format, a...)
}
