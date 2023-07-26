package logger

import (
	"fmt"
	zaprotatelogs "github.com/lestrrat-go/file-rotatelogs"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"os"
	"path"
	"time"
)

type Log struct {
	*zap.Logger
}

var level zapcore.Level

type Config struct {
	Level         string `json:"level" yaml:"level"`
	Format        string `json:"format" yaml:"format"`
	Prefix        string `json:"prefix" yaml:"prefix"`
	Director      string `json:"director"  yaml:"director"`
	LinkName      string `json:"linkName" yaml:"link-name"`
	ShowLine      bool   `json:"showLine" yaml:"showLine"`
	EncodeLevel   string `json:"encodeLevel" yaml:"encode-level"`
	StacktraceKey string `json:"stacktraceKey" yaml:"stacktrace-key"`
	LogInConsole  bool   `json:"logInConsole" yaml:"log-in-console"`
}

func New(isProd bool, optionCfg ...Config) (log *Log) {
	var cfg Config
	if len(optionCfg) > 0 {
		cfg = optionCfg[0]
	} else {
		cfg = Config{
			Level:         "debug",
			Format:        "console",
			Prefix:        "[geeluo]",
			Director:      "log",
			LinkName:      "latest.log",
			ShowLine:      true,
			EncodeLevel:   "LowercaseColorLevelEncoder",
			StacktraceKey: "stacktrace",
			LogInConsole:  !isProd,
		}
	}
	var logger *zap.Logger
	if ok, _ := pathExists(cfg.Director); !ok { // 判断是否有Director文件夹
		fmt.Printf("create %v directory\n", cfg.Director)
		_ = os.MkdirAll(cfg.Director, os.ModePerm)
	}

	switch cfg.Level { // 初始化配置文件的Level
	case "debug":
		level = zap.DebugLevel
	case "info":
		level = zap.InfoLevel
	case "warn":
		level = zap.WarnLevel
	case "error":
		level = zap.ErrorLevel
	case "dpanic":
		level = zap.DPanicLevel
	case "panic":
		level = zap.PanicLevel
	case "fatal":
		level = zap.FatalLevel
	default:
		level = zap.InfoLevel
	}

	if level == zap.DebugLevel || level == zap.ErrorLevel {
		logger = zap.New(getEncoderCore(&cfg), zap.AddStacktrace(level))
	} else {
		logger = zap.New(getEncoderCore(&cfg))
	}
	if cfg.ShowLine {
		logger = logger.WithOptions(zap.AddCaller())
	}

	return &Log{
		logger,
	}
}

// getEncoderConfig 获取zapcore.EncoderConfig
func getEncoderConfig(cfg *Config) (config zapcore.EncoderConfig) {
	config = zapcore.EncoderConfig{
		MessageKey:     "message",
		LevelKey:       "level",
		TimeKey:        "time",
		NameKey:        "logger",
		CallerKey:      "caller",
		StacktraceKey:  config.StacktraceKey,
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.LowercaseLevelEncoder,
		EncodeTime:     customTimeEncoder,
		EncodeDuration: zapcore.SecondsDurationEncoder,
		EncodeCaller:   zapcore.FullCallerEncoder,
	}
	switch {
	case cfg.EncodeLevel == "LowercaseLevelEncoder": // 小写编码器(默认)
		config.EncodeLevel = zapcore.LowercaseLevelEncoder
	case cfg.EncodeLevel == "LowercaseColorLevelEncoder": // 小写编码器带颜色
		config.EncodeLevel = zapcore.LowercaseColorLevelEncoder
	case cfg.EncodeLevel == "CapitalLevelEncoder": // 大写编码器
		config.EncodeLevel = zapcore.CapitalLevelEncoder
	case cfg.EncodeLevel == "CapitalColorLevelEncoder": // 大写编码器带颜色
		config.EncodeLevel = zapcore.CapitalColorLevelEncoder
	default:
		config.EncodeLevel = zapcore.LowercaseLevelEncoder
	}
	return config
}

// getEncoder 获取zapcore.Encoder
func getEncoder(cfg *Config) zapcore.Encoder {
	if cfg.Format == "json" {
		return zapcore.NewJSONEncoder(getEncoderConfig(cfg))
	}
	return zapcore.NewConsoleEncoder(getEncoderConfig(cfg))
}

// getEncoderCore 获取Encoder的zapcore.Core
func getEncoderCore(cfg *Config) (core zapcore.Core) {
	writer, err := getWriteSyncer(cfg) // 使用file-rotatelogs进行日志分割
	if err != nil {
		fmt.Printf("Get Write Syncer Failed err:%v", err.Error())
		return
	}
	return zapcore.NewCore(getEncoder(cfg), writer, level)
}

// 自定义日志输出时间格式
func customTimeEncoder(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
	enc.AppendString(t.Format("[geeluo]" + "2006/01/02 - 15:04:05.000"))
}

func pathExists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}

func getWriteSyncer(cfg *Config) (zapcore.WriteSyncer, error) {
	fileWriter, err := zaprotatelogs.New(
		path.Join(cfg.Director, "%Y-%m-%d.log"),
		zaprotatelogs.WithLinkName(cfg.LinkName),
		zaprotatelogs.WithMaxAge(7*24*time.Hour),
		zaprotatelogs.WithRotationTime(24*time.Hour),
	)
	if cfg.LogInConsole {
		return zapcore.NewMultiWriteSyncer(zapcore.AddSync(os.Stdout), zapcore.AddSync(fileWriter)), err
	}
	return zapcore.AddSync(fileWriter), err
}
