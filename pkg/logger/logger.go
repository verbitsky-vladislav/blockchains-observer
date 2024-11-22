package logger

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type Logger struct {
	serviceName string
	zapLogger   *zap.Logger
}

// NewLogger - инициализация логгера
func NewLogger(serviceName string) (*Logger, error) {
	config := zap.Config{
		Encoding:         "console",
		Level:            zap.NewAtomicLevelAt(zap.InfoLevel),
		OutputPaths:      []string{"stdout"},
		ErrorOutputPaths: []string{"stderr"},
		EncoderConfig: zapcore.EncoderConfig{
			TimeKey:      "timestamp",
			LevelKey:     "level",
			MessageKey:   "message",
			EncodeTime:   zapcore.ISO8601TimeEncoder,
			EncodeLevel:  zapcore.LowercaseLevelEncoder,
			EncodeCaller: zapcore.ShortCallerEncoder,
		},
	}
	zapLogger, err := config.Build()
	if err != nil {
		return nil, err
	}

	return &Logger{
		serviceName: serviceName,
		zapLogger:   zapLogger,
	}, nil
}

// Info - логирование с уровнем Info
func (l *Logger) Info(message string, context interface{}) {
	l.log(zap.InfoLevel, message, context)
}

// Error - логирование с уровнем Error
func (l *Logger) Error(message string, context interface{}) {
	l.log(zap.ErrorLevel, message, context)
}

// Debug - логирование с уровнем Debug
func (l *Logger) Debug(message string, context interface{}) {
	l.log(zap.DebugLevel, message, context)
}

// Panic - логирование с уровнем Panic
func (l *Logger) Panic(message string, context interface{}) {
	l.log(zap.PanicLevel, message, context)
	panic(message)
}

// log - общий метод для логирования
func (l *Logger) log(level zapcore.Level, message string, context interface{}) {
	entry := l.zapLogger.With(
		zap.String("service", l.serviceName),
		zap.Any("context", context),
	)

	switch level {
	case zap.InfoLevel:
		entry.Info(message)
	case zap.ErrorLevel:
		entry.Error(message)
	case zap.DebugLevel:
		entry.Debug(message)
	case zap.PanicLevel:
		entry.Panic(message)
	}
}
