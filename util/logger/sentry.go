// Package logger is simply logger with sentry
package logger

import (
	"context"
	"fmt"
	"log"
	"os"
	"runtime"
	"strings"
	"time"

	"github.com/getsentry/sentry-go"
)

var (
	noLogger       = log.New(os.Stdout, "[NOLEVEL] ", log.LstdFlags|log.Llongfile)
	panicLogger    = log.New(os.Stdout, "[PANIC] ", log.LstdFlags|log.Llongfile)
	creticalLogger = log.New(os.Stdout, "[CRETICAL] ", log.LstdFlags|log.Llongfile)
	errLogger      = log.New(os.Stdout, "[ERROR] ", log.LstdFlags|log.Llongfile)
	warnLogger     = log.New(os.Stdout, "[WARN] ", log.LstdFlags|log.Llongfile)
	infoLogger     = log.New(os.Stdout, "[INFO] ", log.LstdFlags|log.Llongfile)
	debugLogger    = log.New(os.Stdout, "[DEBUG] ", log.LstdFlags|log.Llongfile)
	todoLogger     = log.New(os.Stdout, "[TODO] ", log.LstdFlags|log.Llongfile)
	isDebug        = false
	isSentry       = false
	without        = false
)

// TODO: tidy up
// func trace(deps int) *sentry.Stacktrace {
// 	return sentry.NewStacktrace() // TODO: filter deps, 5, vers
// }

type (
	// MessageFunc uses both func & method(hub.XXXX, sentry.XXXX)
	MessageFunc func(message string) *sentry.EventID

	// NOLEVEL Just for rename in sentry dashboard eventlog title
	NOLEVEL struct{ s string }
	// PANIC Just for rename in sentry dashboard eventlog title
	PANIC struct{ s string }
	// CRETICAL Just for rename in sentry dashboard eventlog title
	CRETICAL struct{ s string }
	// ERROR Just for rename in sentry dashboard eventlog title
	ERROR struct{ s string }
	// WARN Just for rename in sentry dashboard eventlog title
	WARN struct{ s string }
	// INFO Just for rename in sentry dashboard eventlog title
	INFO struct{ s string }
	// DEBUG Just for rename in sentry dashboard eventlog title
	DEBUG struct{ s string }
	// TODO Just for rename in sentry dashboard eventlog title
	TODO struct{ s string }
)

func (e *NOLEVEL) Error() string  { return e.s }
func (e *PANIC) Error() string    { return e.s }
func (e *CRETICAL) Error() string { return e.s }
func (e *ERROR) Error() string    { return e.s }
func (e *WARN) Error() string     { return e.s }
func (e *INFO) Error() string     { return e.s }
func (e *DEBUG) Error() string    { return e.s }
func (e *TODO) Error() string     { return e.s }

// C is alias with cretical
func C(format string, args ...interface{}) { creticaldeps(sentry.CaptureMessage, 3, format, args...) }

// E is alias with error
func E(format string, args ...interface{}) { errdeps(sentry.CaptureMessage, 3, format, args...) }

// W is alias with warning
func W(format string, args ...interface{}) { warndeps(sentry.CaptureMessage, 3, format, args...) }

// I is alias with info
func I(format string, args ...interface{}) { infodeps(sentry.CaptureMessage, 3, format, args...) }

// D is alias with debug
func D(format string, args ...interface{}) { debugdeps(sentry.CaptureMessage, 3, format, args...) }

// T is alias with todo
func T(format string, args ...interface{}) { tododeps(sentry.CaptureMessage, 3, format, args...) }

// CCtx pritns as information
func CCtx(ctx context.Context, format string, args ...interface{}) {
	if hub := sentry.GetHubFromContext(ctx); hub != nil {
		creticaldeps(hub.CaptureMessage, 3, format, args...)
		return
	}

	creticaldeps(sentry.CaptureMessage, 3, format, args...)
}

// ECtx pritns as information
func ECtx(ctx context.Context, format string, args ...interface{}) {
	if hub := sentry.GetHubFromContext(ctx); hub != nil {
		errdeps(hub.CaptureMessage, 3, format, args...)
		return
	}

	errdeps(sentry.CaptureMessage, 3, format, args...)
}

// WCtx pritns as information
func WCtx(ctx context.Context, format string, args ...interface{}) {
	if hub := sentry.GetHubFromContext(ctx); hub != nil {
		warndeps(hub.CaptureMessage, 3, format, args...)
		return
	}

	warndeps(sentry.CaptureMessage, 3, format, args...)
}

// ICtx pritns as information
func ICtx(ctx context.Context, format string, args ...interface{}) {
	if hub := sentry.GetHubFromContext(ctx); hub != nil {
		infodeps(hub.CaptureMessage, 3, format, args...)
		return
	}

	infodeps(sentry.CaptureMessage, 3, format, args...)
}

// DCtx pritns as information
func DCtx(ctx context.Context, format string, args ...interface{}) {
	if hub := sentry.GetHubFromContext(ctx); hub != nil {
		debugdeps(hub.CaptureMessage, 3, format, args...)
		return
	}

	debugdeps(sentry.CaptureMessage, 3, format, args...)
}

// TCtx pritns as information
func TCtx(ctx context.Context, format string, args ...interface{}) {
	if hub := sentry.GetHubFromContext(ctx); hub != nil {
		tododeps(hub.CaptureMessage, 3, format, args...)
		return
	}

	tododeps(sentry.CaptureMessage, 3, format, args...)
}

// SetDebug set debug flag
func SetDebug(debug bool) {
	isDebug = debug
}

// SetSentry set
func SetSentry(sentry bool) {
	isSentry = sentry
}

// TodofWithContext outputs ...
func TodofWithContext(ctx context.Context, format string, args ...interface{}) {
	if hub := sentry.GetHubFromContext(ctx); hub != nil {
		tododeps(hub.CaptureMessage, 3, format, args...)
		return
	}

	tododeps(sentry.CaptureMessage, 3, format, args...)
}

// TodolnWithContext outputs ...
func TodolnWithContext(ctx context.Context, args ...interface{}) {
	s := fmt.Sprintln(args...)

	if hub := sentry.GetHubFromContext(ctx); hub != nil {
		tododeps(hub.CaptureMessage, 3, s)
		return
	}

	tododeps(sentry.CaptureMessage, 3, s)
}

// Todof outputs ...
func Todof(format string, args ...interface{}) {
	tododeps(sentry.CaptureMessage, 3, format, args...)
}

// Todoln outputs ...
func Todoln(args ...interface{}) {
	tododeps(sentry.CaptureMessage, 3, fmt.Sprintln(args...))
}

func tododeps(fn MessageFunc, deps int, format string, args ...interface{}) {
	s := fmt.Sprintf(format, args...)
	_ = todoLogger.Output(deps, s)

	if isSentry {
		_, f, line, _ := runtime.Caller(deps - 1)
		s = fmt.Sprintf("[TODO] %s \non %s:%d", s, f, line)
		// raven.CaptureMessage(s, nil, raven.NewException(&TODO{s}, trace(deps)))
		fn(s)
	}
}

// DebugfWithContext outputs ...
func DebugfWithContext(ctx context.Context, format string, args ...interface{}) {
	if hub := sentry.GetHubFromContext(ctx); hub != nil {
		debugdeps(hub.CaptureMessage, 3, format, args...)
		return
	}

	debugdeps(sentry.CaptureMessage, 3, format, args...)
}

// DebuglnWithContext outputs ...
func DebuglnWithContext(ctx context.Context, args ...interface{}) {
	s := fmt.Sprintln(args...)

	if hub := sentry.GetHubFromContext(ctx); hub != nil {
		debugdeps(hub.CaptureMessage, 3, s)
		return
	}

	debugdeps(sentry.CaptureMessage, 3, s)
}

// Debugf outputs ...
func Debugf(format string, args ...interface{}) {
	debugdeps(sentry.CaptureMessage, 3, format, args...)
}

// Debugln outputs ...
func Debugln(args ...interface{}) {
	debugdeps(sentry.CaptureMessage, 3, fmt.Sprintln(args...))
}

func debugdeps(fn MessageFunc, deps int, format string, args ...interface{}) {
	if isDebug {
		s := fmt.Sprintf(format, args...)
		_ = debugLogger.Output(deps, s)

		if without && isSentry {
			_, f, line, _ := runtime.Caller(deps - 1)
			s = fmt.Sprintf("[DEBUG] %s \non %s:%d", s, f, line)
			// TODO: tidy up
			// raven.CaptureMessage(s, nil, sentry.NewException(&DEBUG{s}, trace(deps)))
			fn(s)
		}
	}
}

// InfofWithContext pritns as information
func InfofWithContext(ctx context.Context, format string, args ...interface{}) {
	if hub := sentry.GetHubFromContext(ctx); hub != nil {
		infodeps(hub.CaptureMessage, 3, format, args...)
		return
	}

	infodeps(sentry.CaptureMessage, 3, format, args...)
}

// InfolnWithContext pritns as information
func InfolnWithContext(ctx context.Context, args ...interface{}) {
	s := fmt.Sprintln(args...)

	if hub := sentry.GetHubFromContext(ctx); hub != nil {
		infodeps(hub.CaptureMessage, 3, s)
		return
	}

	infodeps(sentry.CaptureMessage, 3, s)
}

// Infof pritns as information
func Infof(format string, args ...interface{}) {
	infodeps(sentry.CaptureMessage, 3, format, args...)
}

// Infoln pritns as information
func Infoln(args ...interface{}) {
	infodeps(sentry.CaptureMessage, 3, fmt.Sprintln(args...))
}

func infodeps(fn MessageFunc, deps int, format string, args ...interface{}) {
	s := fmt.Sprintf(format, args...)
	_ = infoLogger.Output(deps, s)

	if without && isSentry {
		_, f, line, _ := runtime.Caller(deps - 1)
		s = fmt.Sprintf("[INFO] %s \non %s:%d", s, f, line)
		// TODO: tidy up
		// raven.CaptureMessage(s, nil, sentry.NewException(&INFO{s}, trace(deps)))
		fn(s)
	}
}

// WarnfWithContext pritns as warning
func WarnfWithContext(ctx context.Context, format string, args ...interface{}) {
	if hub := sentry.GetHubFromContext(ctx); hub != nil {
		warndeps(hub.CaptureMessage, 3, format, args...)
		return
	}

	warndeps(sentry.CaptureMessage, 3, format, args...)
}

// WarningfWithContext pritns as warning
func WarningfWithContext(ctx context.Context, format string, args ...interface{}) {
	if hub := sentry.GetHubFromContext(ctx); hub != nil {
		warndeps(hub.CaptureMessage, 3, format, args...)
		return
	}

	warndeps(sentry.CaptureMessage, 3, format, args...)
}

// WarninglnWithContext outputs ...
func WarninglnWithContext(ctx context.Context, args ...interface{}) {
	s := fmt.Sprintln(args...)

	if hub := sentry.GetHubFromContext(ctx); hub != nil {
		warndeps(hub.CaptureMessage, 3, s)
		return
	}

	warndeps(sentry.CaptureMessage, 3, s)
}

// WarnlnWithContext outputs ...
func WarnlnWithContext(ctx context.Context, args ...interface{}) {
	s := fmt.Sprintln(args...)

	if hub := sentry.GetHubFromContext(ctx); hub != nil {
		warndeps(hub.CaptureMessage, 3, s)
		return
	}

	warndeps(sentry.CaptureMessage, 3, s)
}

// Warnf pritns as warning
func Warnf(format string, args ...interface{}) {
	warndeps(sentry.CaptureMessage, 3, format, args...)
}

// Warningf pritns as warning
func Warningf(format string, args ...interface{}) {
	warndeps(sentry.CaptureMessage, 3, format, args...)
}

// Warningln outputs ...
func Warningln(args ...interface{}) {
	warndeps(sentry.CaptureMessage, 3, fmt.Sprintln(args...))
}

// Warnln outputs ...
func Warnln(args ...interface{}) {
	warndeps(sentry.CaptureMessage, 3, fmt.Sprintln(args...))
}

func warndeps(fn MessageFunc, deps int, format string, args ...interface{}) {
	s := fmt.Sprintf(format, args...)
	_ = warnLogger.Output(deps, s)

	if without && isSentry {
		_, f, line, _ := runtime.Caller(deps - 1)
		s = fmt.Sprintf("[WARN] %s \non %s:%d", s, f, line)
		// TODO: tidy up
		// raven.CaptureMessage(s, nil, sentry.NewException(&WARN{s}, trace(deps)))
		fn(s)
	}
}

// ErrorlnWithContext prints as error
func ErrorlnWithContext(ctx context.Context, args ...interface{}) {
	s := fmt.Sprintln(args...)

	if hub := sentry.GetHubFromContext(ctx); hub != nil {
		errdeps(hub.CaptureMessage, 3, s)
		return
	}

	errdeps(sentry.CaptureMessage, 3, s)
}

// ErrorfWithContext prints as error
func ErrorfWithContext(ctx context.Context, format string, args ...interface{}) {
	if hub := sentry.GetHubFromContext(ctx); hub != nil {
		errdeps(hub.CaptureMessage, 3, format, args...)
		return
	}

	errdeps(sentry.CaptureMessage, 3, format, args...)
}

// ErrfWithContext prints as error
func ErrfWithContext(ctx context.Context, format string, args ...interface{}) {
	if hub := sentry.GetHubFromContext(ctx); hub != nil {
		errdeps(hub.CaptureMessage, 3, format, args...)
		return
	}

	errdeps(sentry.CaptureMessage, 3, format, args...)
}

// ErrlnWithContext outputs ...
func ErrlnWithContext(ctx context.Context, args ...interface{}) {
	s := fmt.Sprintln(args...)

	if hub := sentry.GetHubFromContext(ctx); hub != nil {
		errdeps(hub.CaptureMessage, 3, s)
		return
	}

	errdeps(sentry.CaptureMessage, 3, s)
}

// Errorln pritns as error
func Errorln(args ...interface{}) {
	errdeps(sentry.CaptureMessage, 3, fmt.Sprintln(args...))
}

// Errorf pritns as error
func Errorf(format string, args ...interface{}) {
	errdeps(sentry.CaptureMessage, 3, format, args...)
}

// Errf pritns as error
func Errf(format string, args ...interface{}) {
	errdeps(sentry.CaptureMessage, 3, format, args...)
}

// Errln outputs ...
func Errln(args ...interface{}) {
	errdeps(sentry.CaptureMessage, 3, fmt.Sprintln(args...))
}

func errdeps(fn MessageFunc, deps int, format string, args ...interface{}) {
	s := fmt.Sprintf(format, args...)
	_ = errLogger.Output(deps, s)

	if isSentry {
		_, f, line, _ := runtime.Caller(deps - 1)
		s = fmt.Sprintf("[ERROR] %s \non %s:%d", s, f, line)
		// TODO: tidy up
		// raven.CaptureMessage(s, nil, sentry.NewException(&ERROR{s}, trace(deps)))
		fn(s)
	}
}

// CreticalfWithContext prints as cretical
func CreticalfWithContext(ctx context.Context, format string, args ...interface{}) {
	if hub := sentry.GetHubFromContext(ctx); hub != nil {
		creticaldeps(hub.CaptureMessage, 3, format, args...)
		return
	}

	creticaldeps(sentry.CaptureMessage, 3, format, args...)
}

// CreticalnWithContext outputs ...
func CreticalnWithContext(ctx context.Context, args ...interface{}) {
	s := fmt.Sprintln(args...)

	if hub := sentry.GetHubFromContext(ctx); hub != nil {
		creticaldeps(hub.CaptureMessage, 3, s)
		return
	}

	creticaldeps(sentry.CaptureMessage, 3, s)
}

// CrtlnWithContext prints as cretical
func CrtlnWithContext(ctx context.Context, args ...interface{}) {
	s := fmt.Sprintln(args...)

	if hub := sentry.GetHubFromContext(ctx); hub != nil {
		creticaldeps(hub.CaptureMessage, 3, s)
		return
	}

	creticaldeps(sentry.CaptureMessage, 3, s)
}

// CrtlfWithContext prints as cretical
func CrtlfWithContext(ctx context.Context, format string, args ...interface{}) {
	if hub := sentry.GetHubFromContext(ctx); hub != nil {
		creticaldeps(hub.CaptureMessage, 3, format, args...)
		return
	}

	creticaldeps(sentry.CaptureMessage, 3, format, args...)
}

// Creticalf pritns as cretical
func Creticalf(format string, args ...interface{}) {
	creticaldeps(sentry.CaptureMessage, 3, format, args...)
}

// Creticaln outputs ...
func Creticaln(args ...interface{}) {
	creticaldeps(sentry.CaptureMessage, 3, fmt.Sprintln(args...))
}

// Crtln pritns as cretical
func Crtln(args ...interface{}) {
	creticaldeps(sentry.CaptureMessage, 3, fmt.Sprintln(args...))
}

// Crtlf pritns as cretical
func Crtlf(format string, args ...interface{}) {
	creticaldeps(sentry.CaptureMessage, 3, format, args...)
}

func creticaldeps(fn MessageFunc, deps int, format string, args ...interface{}) {
	s := fmt.Sprintf(format, args...)
	_ = creticalLogger.Output(deps, s)

	if isSentry {
		_, f, line, _ := runtime.Caller(deps - 1)
		s = fmt.Sprintf("[CRETICAL] %s \non %s:%d", s, f, line)
		// TODO: tidy up
		// raven.CaptureMessage(s, nil, sentry.NewException(&CRETICAL{s}, trace(deps)))
		fn(s)
	}
}

// PanicfWithContext prints as panic
func PanicfWithContext(ctx context.Context, format string, args ...interface{}) {
	if hub := sentry.GetHubFromContext(ctx); hub != nil {
		panicdeps(hub.CaptureMessage, 3, format, args...)
		return
	}

	panicdeps(sentry.CaptureMessage, 3, format, args...)
}

// PaniclnWithContext outputs ...
func PaniclnWithContext(ctx context.Context, args ...interface{}) {
	s := fmt.Sprintln(args...)

	if hub := sentry.GetHubFromContext(ctx); hub != nil {
		panicdeps(hub.CaptureMessage, 3, s)
		return
	}

	panicdeps(sentry.CaptureMessage, 3, s)
}

// Panicf pritns as panic
func Panicf(format string, args ...interface{}) {
	panicdeps(sentry.CaptureMessage, 3, format, args...)
}

// Panicln outputs ...
func Panicln(args ...interface{}) {
	panicdeps(sentry.CaptureMessage, 3, fmt.Sprintln(args...))
}

func panicdeps(fn MessageFunc, deps int, format string, args ...interface{}) {
	s := fmt.Sprintf(format, args...)
	_ = panicLogger.Output(deps, s)

	if isSentry {
		_, f, line, _ := runtime.Caller(deps - 1)
		s = fmt.Sprintf("[PANIC] %s \non %s:%d", s, f, line)
		// TODO: tidy up
		// raven.CaptureMessage(s, nil, sentry.NewException(&PANIC{s}, trace(deps)))
		fn(s)
	}

	panic(s)
}

// PrintfWithContext prints with format
func PrintfWithContext(ctx context.Context, format string, args ...interface{}) {
	if hub := sentry.GetHubFromContext(ctx); hub != nil {
		printdeps(hub.CaptureMessage, 3, fmt.Sprintf(format, args...))
		return
	}

	printdeps(sentry.CaptureMessage, 3, fmt.Sprintf(format, args...))
}

// PrintlnWithContext outputs ...
func PrintlnWithContext(ctx context.Context, args ...interface{}) {
	s := fmt.Sprintln(args...)

	if hub := sentry.GetHubFromContext(ctx); hub != nil {
		printdeps(hub.CaptureMessage, 3, s)
		return
	}

	printdeps(sentry.CaptureMessage, 3, s)
}

// Printf pritns with format
func Printf(format string, args ...interface{}) {
	printdeps(sentry.CaptureMessage, 3, fmt.Sprintf(format, args...))
}

// Println outputs ...
func Println(args ...interface{}) {
	printdeps(sentry.CaptureMessage, 3, args...)
}

func printdeps(fn MessageFunc, deps int, args ...interface{}) {
	s := fmt.Sprintln(args...)

	switch {
	case strings.Contains(s, "[PANIC]"):
		creticaldeps(fn, deps+1, s) // TODO: soft panic logging
	case strings.Contains(s, "[CRETICAL]"):
		creticaldeps(fn, deps+1, s)
	case strings.Contains(s, "[ERROR]"):
		errdeps(fn, deps+1, s)
	case strings.Contains(s, "[WARN]"):
		warndeps(fn, deps+1, s)
	case strings.Contains(s, "[INFO]"):
		infodeps(fn, deps+1, s)
	case strings.Contains(s, "[DEBUG]"):
		debugdeps(fn, deps+1, s)
	case strings.Contains(s, "[TODO]"):
		tododeps(fn, deps+1, s)
	default:
		_ = noLogger.Output(deps, s)
	}
}

// Flush waits blocks and waits for all events to finish being sent to Sentry server
func Flush() {
	sentry.Flush(3 * time.Second)
}

// FlushTimeout waits blocks and waits for all events to finish being sent to Sentry server
func FlushTimeout(timeout time.Duration) {
	sentry.Flush(timeout)
}
