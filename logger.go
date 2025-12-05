package gsql

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"runtime"
	"strconv"
	"strings"
	"time"

	"gorm.io/gorm/logger"
)

// LogLevel log level
type LogLevel int

const (
	// LogLevelSilent silent log level
	LogLevelSilent LogLevel = iota + 1
	// LogLevelError error log level
	LogLevelError
	// LogLevelWarn warn log level
	LogLevelWarn
	// LogLevelInfo info log level
	LogLevelInfo
)

type sqlLogger struct {
	logger.Writer
	logger.Config
	infoStr, warnStr, errStr, traceErrStr, traceWarnStr, traceStr string
}

var DefaultLogger = NewLogger(log.New(os.Stdout, "\r\n", log.LstdFlags), logger.Config{
	SlowThreshold:             200 * time.Millisecond,
	LogLevel:                  logger.Warn,
	IgnoreRecordNotFoundError: false,
	Colorful:                  true,
})

// NewWrapperLogger
// Deprecated: Use `NewLogger` instead
func NewWrapperLogger(l logger.Writer, cfg logger.Config) logger.Interface {
	return NewLogger(l, cfg)
}

func NewLogger(l logger.Writer, cfg logger.Config) logger.Interface {
	lg := &sqlLogger{
		Writer: l,
		Config: cfg,
	}
	var (
		infoStr      = "%s\n[info] "
		warnStr      = "%s\n[warn] "
		errStr       = "%s\n[error] "
		traceStr     = "%s\n[%.3fms] [rows:%v] %s"
		traceWarnStr = "%s %s\n[%.3fms] [rows:%v] %s"
		traceErrStr  = "%s %s\n[%.3fms] [rows:%v] %s"
	)

	if lg.Colorful {
		infoStr = logger.Green + "%s\n" + logger.Reset + logger.Green + "[info] " + logger.Reset
		warnStr = logger.BlueBold + "%s\n" + logger.Reset + logger.Magenta + "[warn] " + logger.Reset
		errStr = logger.Magenta + "%s\n" + logger.Reset + logger.Red + "[error] " + logger.Reset
		traceStr = logger.Green + "%s\n" + logger.Reset + logger.Yellow + "[%.3fms] " + logger.BlueBold + "[rows:%v]" + logger.Reset + " %s"
		traceWarnStr = logger.Green + "%s " + logger.Yellow + "%s\n" + logger.Reset + logger.RedBold + "[%.3fms] " + logger.Yellow + "[rows:%v]" + logger.Magenta + " %s" + logger.Reset
		traceErrStr = logger.RedBold + "%s " + logger.MagentaBold + "%s\n" + logger.Reset + logger.Yellow + "[%.3fms] " + logger.BlueBold + "[rows:%v]" + logger.Reset + " %s"
	}
	lg.infoStr = infoStr
	lg.warnStr = warnStr
	lg.errStr = errStr
	lg.traceErrStr = traceErrStr
	lg.traceWarnStr = traceWarnStr
	lg.traceStr = traceStr
	return lg
}

func (l *sqlLogger) LogMode(level logger.LogLevel) logger.Interface {
	newlogger := *l
	newlogger.LogLevel = level
	return &newlogger
}

func (l *sqlLogger) Info(ctx context.Context, msg string, data ...any) {
	if l.LogLevel >= logger.Info {
		l.Printf(l.infoStr+msg, append([]any{FileWithLineNum()}, data...)...)
	}
}

func (l *sqlLogger) Warn(ctx context.Context, msg string, data ...any) {
	if l.LogLevel >= logger.Warn {
		l.Printf(l.warnStr+msg, append([]any{FileWithLineNum()}, data...)...)
	}
}

func (l *sqlLogger) Error(ctx context.Context, msg string, data ...any) {
	if l.LogLevel >= logger.Error {
		l.Printf(l.errStr+msg, append([]any{FileWithLineNum()}, data...)...)
	}
}

func (l *sqlLogger) Trace(ctx context.Context, begin time.Time, fc func() (string, int64), err error) {
	if l.LogLevel <= logger.Silent {
		return
	}

	elapsed := time.Since(begin)
	sql, rows := fc()
	elapsedMs := float64(elapsed.Nanoseconds()) / 1e6
	fileLine := FileWithLineNum()
	rowsValue := any(rows)
	if rows == -1 {
		rowsValue = "-"
	}

	switch {
	case err != nil && l.LogLevel >= logger.Error && (!errors.Is(err, logger.ErrRecordNotFound) || !l.IgnoreRecordNotFoundError):
		l.Printf(l.traceErrStr, fileLine, err, elapsedMs, rowsValue, sql)
	case elapsed > l.SlowThreshold && l.SlowThreshold != 0 && l.LogLevel >= logger.Warn:
		slowLog := fmt.Sprintf("SLOW SQL >= %v", l.SlowThreshold)
		l.Printf(l.traceWarnStr, fileLine, slowLog, elapsedMs, rowsValue, sql)
	case l.LogLevel == logger.Info:
		l.Printf(l.traceStr, fileLine, elapsedMs, rowsValue, sql)
	}
}

// FileWithLineNum return the file name and line number of the current file
func FileWithLineNum() string {
	pcs := [13]uintptr{}
	// the third caller usually from gorm internal
	length := runtime.Callers(1, pcs[:])
	frames := runtime.CallersFrames(pcs[:length])
OUT:
	for i := 0; i < length; i++ {
		// second return value is "more", not "ok"
		frame, _ := frames.Next()
		if strings.Contains(frame.File, "gorm.io") {
			continue
		}
		for _, p := range prefix {
			if strings.HasSuffix(frame.File, p) {
				continue OUT
			}
		}
		return string(strconv.AppendInt(append([]byte(frame.File), ':'), int64(frame.Line), 10))
	}
	return ""
}

var prefix = []string{
	"gsql/query.go",
	"gsql/query_generic.go",
	"gsql/logger.go",
	"gsql/utils.go",
}
