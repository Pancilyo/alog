package alog

import (
	"errors"
	"fmt"
	"os"
	"os/signal"
	"path"
	"runtime"
	"strings"
	"syscall"
	"time"
)

// Level 定义日志等级类型
type Level uint16

// 日志等级常量
const (
	UNKNOWN Level = iota // 未知
	DEBUG                // 调试
	INFO                 // 信息
	WARN                 // 警告
	ERROR                // 错误
	FATAL                // 致命错误
)

// 默认变量
const (
	defaultLevel       = DEBUG
	defaultTimeFormat  = "2006-01-02 15:04:05"
	defaultFilePath    = "./log/"
	defaultFileName    = "ALog.log"
	defaultMaxFileSize = 1 << 23
	defaultMaxChanSize = 50000
	defaultDuration    = time.Hour * 24
)

// Logger 日志接口
type Logger interface {
	Debug(format string, a ...interface{})
	Info(format string, a ...interface{})
	Warn(format string, a ...interface{})
	Error(format string, a ...interface{})
	Fatal(format string, a ...interface{})
}

// Log 日志对象
type Log struct {
	CLogger *consoleLogger // 终端日志输出器
	FLogger *fileLogger    // 文件日志输出器
	isClose bool           // 日志对象是否被关闭
}

// New 构造日志对象,
// 默认类型为 DEBUG
// 默认模式为 控制台模式
func New() *Log {
	logger := &Log{
		CLogger: newConsoleLogger(),
		FLogger: nil,
		isClose: false,
	}
	ch := make(chan os.Signal)
	signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM,
		syscall.SIGKILL, syscall.SIGQUIT)
	go func() {
		<-ch //阻塞程序运行，直到收到终止的信号
		fmt.Println("Cleaning before stop...")
		logger.isClose = true
		// 这里得停一下 让所有的进程都停止输出日志
		time.Sleep(time.Millisecond * 500)
		logger.Close()
		os.Exit(1)
	}()
	return logger
}

// SetConsoleMode 设置仅控制台输出模型
func (l *Log) SetConsoleMode() *Log {
	l.CLogger = newConsoleLogger()
	l.FLogger = nil
	return l
}

// SetFileMode 设置仅文件输出模式
func (l *Log) SetFileMode() *Log {
	l.CLogger = nil
	l.FLogger = newFileLogger()
	return l
}

// SetBothMode 设置控制台、文件双输出模式
func (l *Log) SetBothMode() *Log {
	l.CLogger = newConsoleLogger()
	l.FLogger = newFileLogger()
	return l
}

// SetFilePath 设置输出文件的路径
func (l *Log) SetFilePath(path string) *Log {
	if l.FLogger != nil {
		l.FLogger.filePath = path
	}
	return l
}

// SetFilePath 设置输出文件的名字
func (l *Log) SetFileName(name string) *Log {
	if l.FLogger != nil {
		l.FLogger.fileName = name
	}
	return l
}

// SetLevel 设置日志输出的等级
func (l *Log) SetLevel(str string) *Log {
	level, err := s2L(str)
	if l.CLogger != nil {
		l.CLogger.level = level
	}
	if l.FLogger != nil {
		l.FLogger.level = level
	}
	if err != nil {
		panic(err)
	}
	return l
}

// SetTimeFormat 设置日志输出的时间格式
func (l *Log) SetTimeFormat(format string) *Log {
	if l.CLogger != nil {
		l.CLogger.timeFormat = format
	}
	if l.FLogger != nil {
		l.FLogger.timeFormat = format
	}
	return l
}

// SetSplitMode 设置文件切割模式
func (l *Log) SetSplitMode(mode int16) *Log {
	if l.FLogger != nil {
		l.FLogger.splitMode = mode
	}
	return l
}

// SetMaxFileSize 设置切割文件的大小(仅按大小切割文件时生效)
func (l *Log) SetMaxFileSize(size int64) *Log {
	if l.FLogger != nil {
		l.FLogger.maxFileSize = size
	}
	return l
}

// SetSplitDuration 设置切割文件的间隔时间(仅按时间切割文件时生效)
func (l *Log) SetSplitDuration(duration time.Duration) *Log {
	if duration < time.Minute*10 {
		fmt.Println("切割文件时间至少为10分钟")
		os.Exit(1)
	}
	if l.FLogger != nil {
		l.FLogger.duration = duration
		now := time.Now()
		l.FLogger.startTime = now.Add(-time.Duration(now.UnixNano() % duration.Nanoseconds()))
	}
	return l
}

// Close 关闭日志对象
func (l *Log) Close() {
	if l.FLogger != nil {
		l.FLogger.Close()
	}
}

func (l *Log) Debug(format string, a ...interface{}) {
	if l.FLogger != nil && !l.isClose {
		l.FLogger.Debug(format, a...)
	}
	if l.CLogger != nil && !l.isClose {
		l.CLogger.Debug(format, a...)
	}

}
func (l *Log) Info(format string, a ...interface{}) {
	if l.FLogger != nil && !l.isClose {
		l.FLogger.Info(format, a...)
	}
	if l.CLogger != nil && !l.isClose {
		l.CLogger.Info(format, a...)
	}
}
func (l *Log) Warn(format string, a ...interface{}) {
	if l.FLogger != nil && !l.isClose {
		l.FLogger.Warn(format, a...)
	}
	if l.CLogger != nil && !l.isClose {
		l.CLogger.Warn(format, a...)
	}
}
func (l *Log) Error(format string, a ...interface{}) {
	if l.FLogger != nil && !l.isClose {
		l.FLogger.Error(format, a...)
	}
	if l.CLogger != nil && !l.isClose {
		l.CLogger.Error(format, a...)
	}
}
func (l *Log) Fatal(format string, a ...interface{}) {
	if l.FLogger != nil && !l.isClose {
		l.FLogger.Fatal(format, a...)
	}
	if l.CLogger != nil && !l.isClose {
		l.CLogger.Fatal(format, a...)
	}
}

// s2L (string to Level)字符串转LogLevel
func s2L(str string) (Level, error) {
	// 全转小写英文
	str = strings.ToLower(str)
	switch str {
	case "debug":
		return DEBUG, nil
	case "info":
		return INFO, nil
	case "warn":
		return WARN, nil
	case "error":
		return ERROR, nil
	case "fatal":
		return FATAL, nil
	default:
		err := errors.New("无效的日志级别")
		return UNKNOWN, err
	}
}

// l2S (Level to String) LogLevel转字符串
func l2S(lv Level) string {
	switch lv {
	case DEBUG:
		return "DEBUG"
	case INFO:
		return "INFO"
	case WARN:
		return "WARN"
	case ERROR:
		return "ERROR"
	case FATAL:
		return "FATAL"
	default:
		return "UNKNOWN"
	}
}

// traceInfo 追踪输出信息
func traceInfo(skip int) (fileName, funcName string, line int) {
	pc, file, line, ok := runtime.Caller(skip)
	if !ok {
		fmt.Printf("runtime.Caller() failed\n")
		return
	}
	fileName = path.Base(file)
	funcName = runtime.FuncForPC(pc).Name()
	funcName = strings.Split(funcName, ".")[1]
	return
}
