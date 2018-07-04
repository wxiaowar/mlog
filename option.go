package mlog

import (
	"github.com/valyala/bytebufferpool"
)

type Option interface {
	apply(w *MLogWriter)
}

// optionFunc wraps a func so it satisfies the Option interface.
type optionFunc func(w *MLogWriter)

func (f optionFunc) apply(w *MLogWriter) {
	f(w)
}

func DebugModel() Option {
	return optionFunc(func(w *MLogWriter) {
		w.isDebug = true
	})
}


func FileName(fname string, max int64) Option {
	return optionFunc(func(w *MLogWriter) {
		w.fname = fname
		w.max = max
	})
}

func RotateFile(hfunc RotateHFunc) Option {
	return optionFunc(func(w *MLogWriter) {
		w.rotateFile = hfunc
	})
}

func CacheBuf(max int) Option {
	return optionFunc(func(w *MLogWriter) {
		w.chanbuf = make(chan *bytebufferpool.ByteBuffer, max)
	})
}
