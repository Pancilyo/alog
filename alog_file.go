package alog

import (
	"fmt"
	"os"
	"path"
	"time"
)

// 往文件里面写日志相关代码

const (
	SplitNone       int16 = iota // 不分割
	SplitBaseOnSize              // 按大小分割
	SplitBaseOnTime              // 按时间分割
)

// fileLogger 文件输出日志结构
type fileLogger struct {
	level      Level         // 日志等级
	timeFormat string        // 时间输出格式化
	filePath   string        // 日志文件保存的路径
	fileName   string        // 日志文件保存的文件名
	logChan    chan *logMsg  // 日志通道
	closeChan  chan struct{} // 关闭对象channel
	fileObj    *os.File      // 输出文件句柄
	// 分割文件才需要考虑的参数
	splitMode   int16 // 分割文件的模式
	maxFileSize int64 // 限制文件大小
	// 按时间分割文件的参数
	duration  time.Duration // 时间间隔
	startTime time.Time     // 当前时间起点
}

// logMsg 输出日志消息结构
type logMsg struct {
	level     Level  // 日志等级
	msg       string // 日记消息体
	funcName  string // 函数名
	fileName  string // 文件名
	timeStamp string // 时间
	line      int    // 行号
}

// newFileLogger 构造函数
func newFileLogger() *fileLogger {
	now := time.Now()
	st := now.Add(-time.Duration(now.UnixNano() % defaultDuration.Nanoseconds()))
	fl := &fileLogger{
		level:       defaultLevel,                           // 默认Debug模式
		timeFormat:  defaultTimeFormat,                      // 默认时间输出格式
		filePath:    defaultFilePath,                        // 默认输出文件路径为./log/
		fileName:    defaultFileName,                        // 默认输出文件名为ALog.log
		maxFileSize: defaultMaxFileSize,                     // 默认大小为8MB
		logChan:     make(chan *logMsg, defaultMaxChanSize), // 默认通道为5w，可后续设置
		closeChan:   make(chan struct{}),                    // 判断程序是否关闭
		splitMode:   SplitNone,                              // 默认不分割文件大小
		duration:    defaultDuration,                        // 默认按一天分割
		startTime:   st,                                     // 当前时间起点
	}
	fl.initFile() // 按照文件路径和文件名将文件打开
	return fl
}

// initFile 根据指定的日志文件路径和文件名打开日志文件
func (f *fileLogger) initFile() {
	fullFileName := path.Join(f.filePath, f.fileName)
	// 如果路径不存在为路径创建文件夹
	if err := os.MkdirAll(f.filePath, os.ModePerm); err != nil {
		panic("无法为此路径创建文件夹")
	}

	fileObj, err := os.OpenFile(fullFileName, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0777)
	if err != nil {
		panic(err.Error())
	}
	f.fileObj = fileObj
	// 开启1个后台goroutine去往文件写日志文件
	go f.writeLogBackground()
}

// writeLogBackground 后台写日志文件
func (f *fileLogger) writeLogBackground() {
	for {
		// 判断是否要切割文件
		// 按文件大小切割
		if f.splitMode == SplitBaseOnSize && f.checkSize(f.fileObj) {
			newFile, err := f.splitFile(f.fileObj, SplitBaseOnSize) // 日志文件
			if err != nil {
				panic(err)
			}
			f.fileObj = newFile
		}

		select {
		case <-f.closeChan: // 在程序退出前输出完所有的日志内容到文件中
			// 1秒钟内通道没有日志则关闭日志对象
			t := time.NewTimer(time.Second)
			for {
				select {
				case logTmp := <-f.logChan:
					// 将日志写入文件
					f.writeIntoFile(logTmp)
					// 重置定时器
					t.Reset(time.Second)
				case <-t.C:
					return
				}
			}
		case logTmp := <-f.logChan: //取日志输出到文件
			// 将日志写入文件
			f.writeIntoFile(logTmp)
		default:
			// 取不到日志先休息500毫秒
			time.Sleep(time.Millisecond * 500)
		}
	}
}

// checkSize 根据文件大小判断文件是否需要切割
func (f *fileLogger) checkSize(file *os.File) bool {
	fileInfo, err := file.Stat()
	if err != nil {
		fmt.Printf("get file info failed, err:%v\n", err)
		return false
	}
	// 如果当前文件大小大于等于日志文件的最大值
	return fileInfo.Size() >= f.maxFileSize
}

// TODO 后续完善此方法
// checkSize 时间判断文件是否需要切割
func (f *fileLogger) checkTime(nowTime time.Time) bool {
	return nowTime.Sub(f.startTime) >= f.duration
}

// splitFileByTime 根据时间切割的处理函数
func (f *fileLogger) splitFileByTime(logTime time.Time) {
	if f.splitMode == SplitBaseOnTime && f.checkTime(logTime) {
		newFile, err := f.splitFile(f.fileObj, SplitBaseOnTime) // 日志文件
		subTime := logTime.UnixNano() % f.duration.Nanoseconds()
		if subTime/int64(time.Second) == 0 {
			f.startTime = logTime
		} else {
			f.startTime = logTime.Add(time.Duration(subTime))
		}
		f.startTime = logTime
		if err != nil {
			panic(err)
		}
		f.fileObj = newFile
	}
}

// writeIntoFile 将日志写入文件的处理函数
func (f *fileLogger) writeIntoFile(logTmp *logMsg) {
	// 取出日志，查看日志的时间
	// 判断是否按文件时间切割
	logTime, _ := time.ParseInLocation(f.timeFormat, logTmp.timeStamp, time.Local)
	f.splitFileByTime(logTime)
	// 把日志先拼出来
	logInfo := fmt.Sprintf("[%s] [%s] [%s:%s:%d] %s\n", logTmp.timeStamp, l2S(logTmp.level), logTmp.funcName, logTmp.fileName, logTmp.line, logTmp.msg)
	_, _ = fmt.Fprint(f.fileObj, logInfo)
}

// splitFile 切割文件
func (f *fileLogger) splitFile(file *os.File, mode int16) (*os.File, error) {
	// 需要切割日志文件
	var nowStr string
	if mode == SplitBaseOnSize {
		nowStr = time.Now().Format("20060102150405")
	} else if mode == SplitBaseOnTime {
		if f.duration%(time.Hour*24) == 0 {
			nowStr = f.startTime.Format("20060102")
		} else if f.duration%time.Hour == 0 {
			nowStr = f.startTime.Format("2006010215")
		} else if f.duration%time.Minute == 0 {
			nowStr = f.startTime.Format("200601021504")
		} else {
			nowStr = f.startTime.Format("20060102150405")
		}
		//nowStr = f.startTime.Format("20060102150405")
	}
	fileInfo, err := file.Stat()
	if err != nil {
		fmt.Printf("get file info failed, err:%v\n", err)
		return nil, err
	}
	logName := path.Join(f.filePath, fileInfo.Name())       // 拿到当前的日志文件完整路径
	newLogName := fmt.Sprintf("%s.%s.bak", logName, nowStr) // 拼接一个日志文件备份的名字
	// 1. 关闭当前的日志文件
	_ = file.Close()
	// 2. 备份一下 rename xx.log -> xx.log.20220212150405.bak
	_ = os.Rename(logName, newLogName)
	// 3. 打开一个新的日志文件
	fileObj, err := os.OpenFile(logName, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		fmt.Printf("open new log file failed, err:%v\n", err)
		return nil, err
	}
	// 4. 将打开的新日志文件对象赋值给 f.fileObj
	return fileObj, nil
}

// Close 关闭日志对象
func (f *fileLogger) Close() {
	f.closeChan <- struct{}{} // 关闭log goroutine
	_ = f.fileObj.Close()     // 关闭文件句柄
}

// enable 判断是否需要记录该日志
func (f *fileLogger) enable(lv Level) bool {
	return f.level <= lv
}

// log 记录日志的方法
func (f *fileLogger) log(lv Level, format string, a ...interface{}) {
	if f.enable(lv) {
		msg := fmt.Sprintf(format, a...)           // 拼装消息
		now := time.Now()                          // 获取时间
		fileName, funcName, lineNo := traceInfo(4) // 获取输出此信息的文件名函数名行号
		// 先把日志发送到通道中
		logTmp := &logMsg{
			level:     lv,
			msg:       msg,
			funcName:  funcName,
			fileName:  fileName,
			line:      lineNo,
			timeStamp: now.Format(f.timeFormat),
		}
		select {
		case f.logChan <- logTmp:
		default:
			// 把日志丢掉保证不出现阻塞
		}
	}
}
func (f *fileLogger) Debug(format string, a ...interface{}) {
	f.log(DEBUG, format, a...)
}
func (f *fileLogger) Info(format string, a ...interface{}) {
	f.log(INFO, format, a...)
}
func (f *fileLogger) Warn(format string, a ...interface{}) {
	f.log(WARN, format, a...)
}
func (f *fileLogger) Error(format string, a ...interface{}) {
	f.log(ERROR, format, a...)
}
func (f *fileLogger) Fatal(format string, a ...interface{}) {
	f.log(FATAL, format, a...)
}
