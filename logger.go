// Copyright 2022 Dmitry Heraskou
//
// Licensed under the Apache License, Version 2.0 (the "License"): you may
// not use this file except in compliance with the License. You may obtain
// a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS, WITHOUT
// WARRANTIES OR CONDITIONS OF interface{} KIND, either express or implied. See the
// License for the specific language governing permissions and limitations
// under the License.

// Gojenkins is a Jenkins Client in Go, that exposes the jenkins REST api in a more developer friendly way.
package gojenkins

import "log"

// Logger interface used by jenkins for loging
type LeveledLogger interface {
	Debug(format string, a ...interface{})
	Info(format string, a ...interface{})
	Warn(format string, a ...interface{})
	Error(format string, a ...interface{})
}

// Simple implementation using golang log library
type LeveledLoggerImpl struct {
	debug   *log.Logger
	info    *log.Logger
	warning *log.Logger
	err     *log.Logger
}

func NewLeveledLogger(debug, info, warn, err *log.Logger) LeveledLogger {
	return &LeveledLoggerImpl{
		debug:   debug,
		info:    info,
		warning: warn,
		err:     err,
	}
}

func (l *LeveledLoggerImpl) Debug(format string, a ...interface{}) {
	if l == nil || l.debug == nil {
		return
	}
	l.debug.Printf(format, a)
}

func (l *LeveledLoggerImpl) Info(format string, a ...interface{}) {
	if l == nil || l.info == nil {
		return
	}
	l.info.Printf(format, a)
}

func (l *LeveledLoggerImpl) Warn(format string, a ...interface{}) {
	if l == nil || l.warning == nil {
		return
	}
	l.warning.Printf(format, a)
}

func (l *LeveledLoggerImpl) Error(format string, a ...interface{}) {
	if l == nil || l.err == nil {
		return
	}
	l.err.Printf(format, a)
}
