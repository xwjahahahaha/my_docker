package log

import (
	logrus "github.com/sirupsen/logrus"
	"os"
)

func init() {
	// 设置输出为标准输出
	Log.Out = os.Stdout
	// 一些配置
	// 日志输出等级
	Log.SetLevel(logrus.DebugLevel)
	// 输出格式
	Log.SetFormatter(&logrus.TextFormatter{
		DisableColors: false,
		FullTimestamp: true,
	})
	// 是否追踪方法
	Log.SetReportCaller(false)
}