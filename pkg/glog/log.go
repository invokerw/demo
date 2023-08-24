package glog

import "fmt"

type Logger interface {
	With(args ...interface{}) Logger
	Debug(args ...interface{})
	Info(args ...interface{})
	Warn(args ...interface{})
	Error(args ...interface{})
	Debugf(template string, args ...interface{})
	Infof(template string, args ...interface{})
	Warnf(template string, args ...interface{})
	Errorf(template string, args ...interface{})
	SyncLogger() error
}

var _ Logger = (*LoggerImpl)(nil)

type LoggerImpl struct {
}

func (this *LoggerImpl) Debug(args ...interface{}) {
	fmt.Println(args...)
}

func (this *LoggerImpl) Info(args ...interface{}) {
	fmt.Println(args...)
}

func (this *LoggerImpl) Warn(args ...interface{}) {
	fmt.Println(args...)
}

func (this *LoggerImpl) Error(args ...interface{}) {
	fmt.Println(args...)
}

func (this *LoggerImpl) Debugf(template string, args ...interface{}) {
	fmt.Printf(template, args...)
}

func (this *LoggerImpl) Infof(template string, args ...interface{}) {
	fmt.Printf(template, args...)
}

func (this *LoggerImpl) Warnf(template string, args ...interface{}) {
	fmt.Printf(template, args...)
}

func (this *LoggerImpl) Errorf(template string, args ...interface{}) {
	_ = fmt.Errorf(template, args...)
}

func (this *LoggerImpl) SyncLogger() error {
	return nil
}

func (this *LoggerImpl) With(args ...interface{}) Logger {
	return this
}
