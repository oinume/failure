package failure

import (
	"fmt"
	"path"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/pkg/errors"
)

// CallStack represents call stack.
type CallStack []Frame

// Format implements fmt.Formatter.
func (cs CallStack) Format(s fmt.State, verb rune) {
	switch verb {
	case 'v':
		switch {
		case s.Flag('+'):
			for _, pc := range cs {
				fmt.Fprintf(s, "%+v\n", pc)
			}
		case s.Flag('#'):
			fmt.Fprintf(s, "%#v", []Frame(cs))
		default:
			l := len(cs)
			if l == 0 {
				return
			}
			for _, pc := range cs[:l-1] {
				fmt.Fprintf(s, "%s: ", pc.Func())
			}
			fmt.Fprintf(s, "%v", cs[l-1].Func())
		}
	case 's':
		fmt.Fprintf(s, "%v", cs)
	}
}

// Callers returns call stack for the current state.
func Callers(skip int) CallStack {
	var pcs [32]uintptr
	n := runtime.Callers(skip+2, pcs[:])
	if n == 0 {
		return nil
	}

	fs := runtime.CallersFrames(pcs[:n])

	cs := make(CallStack, 0, n)
	for {
		f, more := fs.Next()

		cs = append(cs, frame{f.File, f.Line, f.Function})

		if !more {
			break
		}
	}

	return cs
}

// CallStackFromPkgErrors creates CallStack from errors.StackTrace.
func CallStackFromPkgErrors(st errors.StackTrace) CallStack {
	pcs := make([]uintptr, len(st))
	for i, v := range st {
		pcs[i] = uintptr(v)
	}

	fs := runtime.CallersFrames(pcs)

	cs := make(CallStack, 0, len(pcs))
	for {
		f, more := fs.Next()

		cs = append(cs, frame{f.File, f.Line, f.Function})

		if !more {
			break
		}
	}
	return cs
}

type Frame interface {
	Path() string
	File() string
	Line() int
	Func() string
	Pkg() string
}

// PC represents program counter.
type frame struct {
	file     string
	line     int
	function string
}

// Path returns a full path to the file for pc.
func (f frame) Path() string {
	return f.file
}

// File returns a file name for pc.
func (f frame) File() string {
	return filepath.Base(f.file)
}

// Line returns a line number for pc.
func (f frame) Line() int {
	return f.line
}

// Func returns a function name for pc.
func (f frame) Func() string {
	fs := strings.Split(path.Base(f.function), ".")
	if len(fs) >= 1 {
		return strings.Join(fs[1:], ".")
	}
	return fs[0]
}

// Pkg returns a package name for pc.
func (f frame) Pkg() string {
	fs := strings.Split(path.Base(f.function), ".")
	return fs[0]
}

// Format implements fmt.Formatter.
func (f frame) Format(s fmt.State, verb rune) {
	switch verb {
	case 'v':
		if s.Flag('+') {
			fmt.Fprintf(s, "[%s] ", f.Func())
		}
		fallthrough
	case 's':
		fmt.Fprintf(s, "%s:%d", f.Path(), f.Line())
	}
}
