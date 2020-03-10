package server

import (
	"fmt"
	"io"
	"os"

	"github.com/fastbill/go-service-toolkit/v2/observance"
	"github.com/labstack/gommon/log"
)

// Logger is a wrapper for the logrus fieldlogger that fulfills the echo logger interface
type Logger struct {
	observance.Logger
}

// Output returns the default output.
func (l Logger) Output() io.Writer {
	return os.Stdout
}

// Prefix is NOT IMPLEMENTED, it is only present to fulfill the echo logger interface
func (l Logger) Prefix() string {
	l.Error("trying to use unimplemented log method Prefix")
	return ""
}

// SetPrefix is NOT IMPLEMENTED, it is only present to fulfill the echo logger interface
func (l Logger) SetPrefix(p string) {
	l.Error("trying to use unimplemented log method SetPrefix")
}

// Level is NOT IMPLEMENTED, it is only present to fulfill the echo logger interface
func (l Logger) Level() log.Lvl {
	l.Error("trying to use unimplemented log method Level")
	return 0
}

// SetLevel is NOT IMPLEMENTED, it is only present to fulfill the echo logger interface
func (l Logger) SetLevel(v log.Lvl) {
	l.Error("trying to use unimplemented log method SetLevel")
}

// SetHeader is NOT IMPLEMENTED, it is only present to fulfill the echo logger interface
func (l Logger) SetHeader(h string) {
	l.Error("trying to use unimplemented log method SetHeader")
}

// Panic implements echo.Logger#Panic.
func (l Logger) Panic(i ...interface{}) {
	l.Error(fmt.Sprintf("%v", i))
}

// Panicf implements echo.Logger#Panicf.
func (l Logger) Panicf(format string, args ...interface{}) {
	l.Error(fmt.Sprintf(format, args...))
}

// Panicj implements echo.Logger#Panicj.
func (l Logger) Panicj(j log.JSON) {
	l.Error(fmt.Sprintf("%+v", j))
}

// Fatal implements echo.Logger#Fatal.
func (l Logger) Fatal(i ...interface{}) {
	l.Error(fmt.Sprintf("%v", i))
}

// Fatalf implements echo.Logger#Fatalf.
func (l Logger) Fatalf(format string, args ...interface{}) {
	l.Error(fmt.Sprintf(format, args...))
}

// Fatalj implements echo.Logger#Fatalj.
func (l Logger) Fatalj(j log.JSON) {
	l.Error(fmt.Sprintf("%+v", j))
}

// Error implements echo.Logger#Error.
func (l Logger) Error(i ...interface{}) {
	l.Logger.Error(fmt.Sprintf("%v", i))
}

// Errorf implements echo.Logger#Errorf.
func (l Logger) Errorf(format string, args ...interface{}) {
	l.Error(fmt.Sprintf(format, args...))
}

// Errorj implements echo.Logger#Errorj.
func (l Logger) Errorj(j log.JSON) {
	l.Error(fmt.Sprintf("%+v", j))
}

// Warn implements echo.Logger#Warn.
func (l Logger) Warn(i ...interface{}) {
	l.Logger.Warn(fmt.Sprintf("%v", i))
}

// Warnf implements echo.Logger#Warnf.
func (l Logger) Warnf(format string, args ...interface{}) {
	l.Warn(fmt.Sprintf(format, args...))
}

// Warnj implements echo.Logger#Warnj.
func (l Logger) Warnj(j log.JSON) {
	l.Warn(fmt.Sprintf("%+v", j))
}

// Info implements echo.Logger#Info.
func (l Logger) Info(i ...interface{}) {
	l.Logger.Info(fmt.Sprintf("%v", i))
}

// Infof implements echo.Logger#Infof.
func (l Logger) Infof(format string, args ...interface{}) {
	l.Info(fmt.Sprintf(format, args...))
}

// Infoj implements echo.Logger#Infoj.
func (l Logger) Infoj(j log.JSON) {
	l.Info(fmt.Sprintf("%+v", j))
}

// Debug implements echo.Logger#Debug.
func (l Logger) Debug(i ...interface{}) {
	l.Logger.Debug(fmt.Sprintf("%v", i))
}

// Debugf implements echo.Logger#Debugf.
func (l Logger) Debugf(format string, args ...interface{}) {
	l.Debug(fmt.Sprintf(format, args...))
}

// Debugj implements echo.Logger#Debugj.
func (l Logger) Debugj(j log.JSON) {
	l.Debug(fmt.Sprintf("%+v", j))
}

// Print implements echo.Logger#Print.
func (l Logger) Print(i ...interface{}) {
	l.Info(fmt.Sprintf("%v", i))
}

// Printf implements echo.Logger#Printf.
func (l Logger) Printf(format string, args ...interface{}) {
	l.Info(fmt.Sprintf(format, args...))
}

// Printj implements echo.Logger#Printj.
func (l Logger) Printj(j log.JSON) {
	l.Info(fmt.Sprintf("%+v", j))
}
