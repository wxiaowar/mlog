package mlog

import (
	"encoding/json"
	"fmt"
	"github.com/valyala/bytebufferpool"
	"os"
	"sync"
	"sync/atomic"
	"time"
)

type RotateHFunc func(string) *os.File

type MLogWriter struct {
	*os.File

	fname string
	max   int64
	cur   int64

	stat_t_cnt  int64
	stat_t_size int64
	stat_s_cnt  int64
	stat_s_size int64

	isDebug    bool
	wait       sync.WaitGroup
	chanbuf    chan *bytebufferpool.ByteBuffer
	rotateFile RotateHFunc
}

//
func NewLogWriter(option ...Option) *MLogWriter {
	ml := &MLogWriter{
		max:        int64(1 << 61),
		File:       os.Stdout,
		chanbuf:    make(chan *bytebufferpool.ByteBuffer, 512),
		rotateFile: defaultRotate,
	}

	for _, opt := range option {
		opt.apply(ml)
	}

	go ml.realWrite()
	return ml
}

func (m *MLogWriter) Static() []byte {
	ret := map[string]interface{}{
		"t_cnt":  m.stat_t_cnt,
		"t_size": m.stat_t_size,
		"s_cnt":  m.stat_s_cnt,
		"s_size": m.stat_s_size,
		"cache":  len(m.chanbuf),
	}

	bts, err := json.Marshal(ret)
	if err != nil {
		fmt.Fprintf(os.Stderr, "MLog Static Error %v\n", err)
	}

	return bts
}

// Write Note: should stop loginc before stop log
func (m *MLogWriter) Write(p []byte) (n int, err error) {
	atomic.AddInt64(&m.stat_t_cnt, 1)
	atomic.AddInt64(&m.stat_t_size, int64(len(p)))

	b := bytebufferpool.Get()
	b.Write(p)

	select {
	case m.chanbuf <- b:
	default:
		os.Stderr.Write(p)
		return 0, fmt.Errorf("cache full")
	}

	return
}

func (m *MLogWriter) realWrite() {
	m.wait.Add(1)
	byteCache := NewBuffer(2048)
	m.File = m.rotateFile(m.fname)

	defer func() {
		n, err := byteCache.WriteTo(m.File)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Write [%s] Error %v\n", byteCache.String(), err)
		}

		m.stat_s_size += n

		m.File.Sync()
		if m.File != os.Stderr {
			m.File.Close()
		}

		m.wait.Done()
	}()

	for p := range m.chanbuf {
		if m.isDebug {
			fmt.Fprintf(os.Stdout, "%s", p.String())
		}

		m.stat_s_cnt++
		if byteCache.Write(p.Bytes()) {
			bytebufferpool.Put(p)
			continue
		}

		n, err := byteCache.WriteTo(m.File)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Write[%s] Error %v\n", byteCache.String(), err)
			m.reopen()
		} else {
			m.stat_s_size += n
			m.cur += int64(n)
			if m.cur > m.max { // rotate
				m.reopen()
			}
		}

		byteCache.Write(p.Bytes())
		bytebufferpool.Put(p)
	}
}

func (m *MLogWriter) reopen() {
	m.File.Sync()
	if m.File != os.Stderr {
		m.File.Close()
	}

	m.cur = 0
	m.File = m.rotateFile(m.fname)
}

// Sync stop rountine and sync file
func (m *MLogWriter) Sync() {
	close(m.chanbuf)
	m.wait.Wait()
}

func defaultRotate(fname string) *os.File {
	nfname := fmt.Sprintf("%s_%s", fname, time.Now().Format("2006-01-02T15-04-05"))
	f, err := os.OpenFile(nfname, os.O_RDWR|os.O_APPEND|os.O_CREATE|os.O_TRUNC, 0666)
	if err != nil {
		fmt.Fprintf(os.Stderr, "OpenFile[%v] Error %v", fname, err)
		return os.Stderr
	}

	return f
}
