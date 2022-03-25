# alog日志库

## 描述

一个带有trace的日志库

特性

* 可输出至控制台或文件.
* 支持文件按大小或按时间切割
* 可设置不同的输出等级

## 开始

引用库

```bash
go get -u -v github.com/Pancilyo/alog
```



## 使用样例

```go
logger := alog.New()
logger.Debug("测试的Debug信息")
logger.Info("测试的Info信息")
logger.Warn("测试的Warn信息")
logger.Error("测试的Error信息")
logger.Fatal("测试的Fatal信息")
```

输出样式

```go
[2022-03-25 20:24:24] [DEBUG] [main.go:main:10] 测试的Debug信息
[2022-03-25 20:24:24] [INFO] [main.go:main:11] 测试的Info信息
[2022-03-25 20:24:24] [WARN] [main.go:main:12] 测试的Warn信息
[2022-03-25 20:24:24] [ERROR] [main.go:main:13] 测试的Error信息
[2022-03-25 20:24:24] [FATAL] [main.go:main:14] 测试的Fatal信息
```

## 设置输出模式

```go
logger := alog.New()					// 默认为控制台模式
logger := alog.New().SetFileMode()		// 文件模式
logger := alog.New().SetConsoleMode()	// 控制台模式
logger := alog.New().SetBothMode() 		// 控制台+文件模式
```

## 设置输出文件相关

```go
// 输出路径： ./log/ALog.log 
// 文件按大小分割 (4MB)
logger := alog.New().SetFileMode().
		SetFilePath("./log/").SetFileName("./ALog.log").
		SetSplitMode(alog.SplitBaseOnSize).SetMaxFileSize(1<<22)

// 默认输出路径： ./log/ALog.log 
// 文件按时间分割 (24小时)
logger := alog.New().SetFileMode().
		SetSplitMode(alog.SplitBaseOnTime).SetSplitDuration(time.Hour * 24)
```

