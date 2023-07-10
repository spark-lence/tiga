package tiga

import (
	"fmt"
	"os"
	"strings"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

var Logger *logrus.Logger = logrus.New()
var ProcessId string = uuid.New().String()

type DefaultFieldHook struct {
}

func (hook *DefaultFieldHook) Fire(entry *logrus.Entry) error {
	// u4 := uuid.New()
	name, _ := os.Hostname()
	// entry.Data["uuid"] = u4.String()
	entry.Data["hostname"] = name
	// entry.Data["function"] = entry.Caller.Function
	return nil
}

func (hook *DefaultFieldHook) Levels() []logrus.Level {
	return logrus.AllLevels
}
func GetLogger(Name string) *logrus.Entry {

	log := Logger.WithFields(logrus.Fields{
		"logName": Name,
	})

	return log
}
func InitLog(config Configuration) {
	// 设置路径
	Logger.SetReportCaller(true)

	Logger.SetOutput(os.Stdout)
	_, ex := os.LookupEnv("UNITTEST")
	logLevel := config.GetString("log.level")
	if ex {
		logLevel = "error"
	}
	logLevel = strings.TrimSpace(logLevel)
	if logLevel == "" {
		logLevel = "info"
	}
	level, err := logrus.ParseLevel(logLevel)
	if err != nil {
		panic(fmt.Errorf("fatal error parse level: %s", err))
	}
	Logger.SetFormatter(&logrus.TextFormatter{
		ForceQuote:      true,                  //键值对加引号
		TimestampFormat: "2006-01-02 15:04:05", //时间格式
		FullTimestamp:   true,
	})
	Logger.SetLevel(logrus.Level(level))
	// Logger.Hooks.Add(lfshook.NewHook(pathMap, Logger.Formatter))
	Logger.Hooks.Add(&DefaultFieldHook{})
}
