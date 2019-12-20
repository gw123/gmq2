package core

import (
	"bytes"
	"github.com/gw123/gmq/core/interfaces"
	"github.com/sirupsen/logrus"
	"io"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
)

const (
	DriverLogrus = "logrus"
	DriverSimple = "simple"
	DriverEs     = "es"
)

type LogItem struct {
	CreatedAt string `json:"created_at"`
	Level     string `json:"level"`
	Cate      string `json:"cate"`
	Msg       string `json:"msg"`
}

const EsLogIndex = "gmq-log"

type LogManager struct {
	app                 interfaces.App
	fileHandel          *os.File
	isAsync             bool
	interval            int
	filterLogCategories []string
	onlyLogCategories   []string
	level               int
	buffer              *bytes.Buffer
	mutex               sync.Mutex
	logger              *logrus.Logger
	//logDriver           string
	timestampFormat     string
}

func NewLogManager(app interfaces.App) *LogManager {
	this := new(LogManager)
	this.app = app
	this.buffer = new(bytes.Buffer)
	this.isAsync = true
	this.interval = 1

	filterCategories, _ := this.app.GetAppConfigItem("logFilterCategories")
	logOnlyCategories, _ := this.app.GetAppConfigItem("logOnlyCategories")
	logInterval, _ := this.app.GetAppConfigItem("logInterval")
	this.timestampFormat = this.app.GetConfig().GetString("logger.timestampFormat")
	if this.timestampFormat == "" {
		this.timestampFormat = "2006-01-02 15:04:05"
	}

	if filterCategories != "" {
		this.filterLogCategories = strings.Split(filterCategories, ",")
	}
	if logOnlyCategories != "" {
		this.onlyLogCategories = strings.Split(logOnlyCategories, ",")
	}

	this.interval, _ = strconv.Atoi(logInterval)
	if this.interval < 1 {
		this.interval = 1
	}

	//this.logDriver = this.app.GetConfig().GetString("logger.logDriver")
	this.initLogrus()
	//this.CreatedLogFile()
	return this
}

func (this LogManager) CreatedLogFile() {
	date := time.Now().Format("2006-01-02")
	var err error
	this.fileHandel, err = os.OpenFile(date+".log", os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0660)
	if err != nil {
		this.Error("LogManager", err.Error())
	}
}

func (this *LogManager) initLogrus() {
	cfg := this.app.GetConfig()
	level := cfg.GetString("logger.level")
	forceColors := cfg.GetBool("logger.forceColors")
	formatter := cfg.GetString("logger.formatter")

	logger := logrus.New()

	if formatter == "text" {
		logger.SetFormatter(&logrus.TextFormatter{
			ForceColors:      forceColors,
			FullTimestamp:    true,
			TimestampFormat:  this.timestampFormat,
			QuoteEmptyFields: true,
		})
	} else {
		logger.SetFormatter(&logrus.JSONFormatter{
			TimestampFormat: this.timestampFormat,
		})
	}

	logger.SetOutput(this.buffer)
	lvl, err := logrus.ParseLevel(level)
	if err != nil {
		lvl = logrus.InfoLevel
	}
	logger.SetLevel(lvl)
	this.logger = logger
}

func (this *LogManager) SetIsAsync(flag bool) {
	this.isAsync = flag
}

func (this *LogManager) SetIntVal(interval int) {
	this.interval = interval
}

func (this *LogManager) Info(category string, format string, a ...interface{}) {
	this.filter("Info", category, format, a...)
}

func (this *LogManager) Warn(category string, format string, a ...interface{}) {
	this.filter("Warn", category, format, a...)
}

func (this *LogManager) Error(category string, format string, a ...interface{}) {
	this.filter("Error", category, format, a...)
}

func (this *LogManager) Debug(category string, format string, a ...interface{}) {
	this.filter("Debug", category, format, a...)
}

func (this *LogManager) filter(logLevel, category string, format string, a ...interface{}) {
	var pass = true
	if len(this.onlyLogCategories) != 0 {
		pass = false
		for _, cate := range this.onlyLogCategories {
			if cate == category {
				pass = true
			}
		}
	} else {
		for _, cate := range this.filterLogCategories {
			if cate == category {
				return
			}
		}
	}

	if !pass {
		return
	}

	logLevel = strings.ToLower(logLevel)
	lvl, err := logrus.ParseLevel(logLevel)
	if err != nil {
		lvl = logrus.WarnLevel
	}

	this.logger.WithFields(logrus.Fields{
		"module": category,
	}).Logf(lvl, format, a...)
}

//func (this *LogManager) PushEs(level, category, date, msg string) {
//	item := LogItem{
//		CreatedAt: date,
//		Level:     level,
//		Cate:      category,
//		Msg:       msg,
//	}
//	esService, ok := this.app.GetService("EsService").(*services.EsService)
//	if !ok {
//		return
//	}
//	data, _ := json.Marshal(item)
//
//	es := esService.GetEs()
//	_, err := es.Index(EsLogIndex, bytes.NewReader(data))
//	if err != nil {
//		fmt.Println(err)
//	}
//}

func (this *LogManager) Write(data []byte) (n int, err error) {
	return this.buffer.Write([]byte(data))
}

func (this *LogManager) Start() {
	go func() {
		for ; ; {
			line, err := this.buffer.ReadString('\n')
			if err != nil {
				if err == io.EOF {
					time.Sleep(time.Second * time.Duration(this.interval))
					continue
				}
				logrus.Error("this.buffer.Read(buffer): " + err.Error())
				break
			}
			if line == "" {
				time.Sleep(time.Second * time.Duration(this.interval))
				continue
			}
			//fmt.Print(line)
			//if this.app.GetConfig().GetBool("logger.enable") {
			//	esService, ok := this.app.GetService("EsService").(*services.EsService)
			//	if !ok {
			//		fmt.Println("es 服务错误")
			//		continue
			//	}
			//	es := esService.GetEs()
			//	_, err = es.Index(EsLogIndex, strings.NewReader(line))
			//	if err != nil {
			//		fmt.Println(err)
			//	}
			//}
			//_, err = this.fileHandel.Write(buffer[0:readLen])
			//if err != nil {
			//	logrus.Error("this.fileHandel.Write(buffer[0:readLen]): "+err.Error())
			//}
		}
	}()
}
