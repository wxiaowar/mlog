package mlog

// Option argus for MLog
type Option interface {
	apply(w *writer)
}

// optionFunc wraps a func so it satisfies the Option interface.
type optionFunc func(w *writer)

func (f optionFunc) apply(w *writer) {
	f(w)
}

// FileName set log file name
func Logfile(name string, max int64) Option {
	return optionFunc(func(w *writer) {
		w.name = name
		w.max = max
	})
}

//
func Cache(try, max int, tick int64, alert Alerter) Option {
	return optionFunc(func(w *writer) {
		w.cache = NewRing(try, max, alert)
		w.cacheMax = max / 2
		w.tick = tick
	})
}


