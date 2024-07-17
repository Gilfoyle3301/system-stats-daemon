package logger

import "log/slog"

type Logger interface {
	Debug(msg string)
	Info(msg string)
	Warn(msg string)
	Error(msg string)
}

type Log struct {
	newLoger *slog.Logger
}

func (l *Log) Debug(msg string) {
	l.newLoger.Debug(msg)
}

func (l *Log) Info(msg string) {
	l.newLoger.Info(msg)
}

func (l *Log) Warn(msg string) {
	l.newLoger.Warn(msg)
}

func (l *Log) Error(msg string) {
	l.newLoger.Error(msg)
}
