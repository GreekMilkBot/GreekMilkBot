package log

import (
	"strings"
	"sync"

	"go.uber.org/zap"
	"go.uber.org/zap/buffer"
	"go.uber.org/zap/zapcore"
)

var (
	mu                 sync.Mutex
	loggers            = make(map[string]*zap.Logger)
	defaultAtomicLevel = zap.NewAtomicLevelAt(zap.InfoLevel)
)

const DEFAULT = "default"

func init() {
	_ = zap.RegisterEncoder("space", newSpaceEncoderFactory)
}

func newSpaceEncoderFactory(cfg zapcore.EncoderConfig) (zapcore.Encoder, error) {
	return NewSpaceEncoder(cfg), nil
}

type spaceEncoder struct {
	zapcore.Encoder
}

func NewSpaceEncoder(cfg zapcore.EncoderConfig) zapcore.Encoder {
	return &spaceEncoder{Encoder: zapcore.NewConsoleEncoder(cfg)}
}

func (enc *spaceEncoder) EncodeEntry(entry zapcore.Entry, fields []zapcore.Field) (*buffer.Buffer, error) {
	buf, err := enc.Encoder.EncodeEntry(entry, fields)
	if err != nil {
		return nil, err
	}
	str := strings.ReplaceAll(buf.String(), "\t", " ")
	buf.Reset()
	buf.AppendString(str)
	return buf, nil
}

func newCompactEncoderConfig() zapcore.EncoderConfig {
	return zapcore.EncoderConfig{
		TimeKey:        "timeKey",
		LevelKey:       "levelKey",
		CallerKey:      "",
		MessageKey:     "messageKey",
		StacktraceKey:  "",
		EncodeTime:     zapcore.ISO8601TimeEncoder,
		EncodeLevel:    zapcore.CapitalColorLevelEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
		EncodeDuration: zapcore.StringDurationEncoder,
	}
}

func initLogger(name string) *zap.Logger {
	config := zap.Config{
		Level:            defaultAtomicLevel,
		Development:      false,
		Encoding:         "space",
		EncoderConfig:    newCompactEncoderConfig(),
		OutputPaths:      []string{"stdout"},
		ErrorOutputPaths: []string{"stderr"},
	}
	logger, err := config.Build()
	if err != nil {
		panic("failed to init logger [" + name + "]: " + err.Error())
	}
	return logger
}

func GetLogger(name string) *zap.Logger {
	mu.Lock()
	defer mu.Unlock()

	if logger, exists := loggers[name]; exists {
		return logger
	}

	logger := initLogger(name)
	loggers[name] = logger
	return logger
}

func SetLevel(level zapcore.Level) {
	defaultAtomicLevel.SetLevel(level)
}

func Debugf(format string, values ...any) {
	GetLogger(DEFAULT).Sugar().Debugf(format, values...)
}

func Infof(format string, values ...any) {
	GetLogger(DEFAULT).Sugar().Infof(format, values...)
}

func Warnf(format string, values ...any) {
	GetLogger(DEFAULT).Sugar().Warnf(format, values...)
}

func Errorf(format string, values ...any) {
	GetLogger(DEFAULT).Sugar().Errorf(format, values...)
}

func Fatalf(format string, values ...any) {
	GetLogger(DEFAULT).Sugar().Fatalf(format, values...)
}
