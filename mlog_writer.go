package mlog

import (
	"encoding/json"
	"fmt"
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
	rotateFile RotateHFunc
}

//
func NewLogWriter(option ...Option) *MLogWriter {
	ml := &MLogWriter{
		max:        int64(1 << 61),
		File:       os.Stdout,
		rotateFile: defaultRotate,
	}

	for _, opt := range option {
		opt.apply(ml)
	}

	ml.File = ml.rotateFile(ml.LogName())
	return ml
}

func (m *MLogWriter) Static() []byte {
	ret := map[string]interface{}{
		"t_cnt":  m.stat_t_cnt,
		"t_size": m.stat_t_size,
		"s_cnt":  m.stat_s_cnt,
		"s_size": m.stat_s_size,
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
	if m.isDebug {
		fmt.Fprintf(os.Stdout, "%s", string(p))
	}
	return m.File.Write(p)
}

//func (m *MLogWriter) realWrite(p []byte) {
//	byteCache := NewBuffer(2048)
//	m.File = m.rotateFile(m.fname)
//
//	defer func() {
//		n, err := byteCache.WriteTo(m.File)
//		if err != nil {
//			fmt.Fprintf(os.Stderr, "Write [%s] Error %v\n", byteCache.String(), err)
//		}
//
//		m.stat_s_size += n
//
//		m.File.Sync()
//		if m.File != os.Stderr {
//			m.File.Close()
//		}
//
//		m.wait.Done()
//	}()
//
//		if m.isDebug {
//			fmt.Fprintf(os.Stdout, "%s", string(p))
//		}
//
//		m.stat_s_cnt++
//		n, err := byteCache.WriteTo(m.File)
//		if err != nil {
//			fmt.Fprintf(os.Stderr, "Write[%s] Error %v\n", byteCache.String(), err)
//			m.reopen()
//		} else {
//			m.stat_s_size += n
//			m.cur += int64(n)
//			if m.cur > m.max { // rotate
//				m.reopen()
//			}
//		}
//
//		byteCache.Write(p)
//}

func (m MLogWriter) LogName()string {
	return m.fname + ".log"
}

func (m *MLogWriter) reopen() {
	m.File.Sync()
	if m.File != os.Stderr {
		m.File.Close()
	}

	lname := m.LogName()
	name := fmt.Sprintf("%s_%s.log", m.fname, time.Now().Format("2006-01-02T15"))
	os.Rename(lname, name)

	m.cur = 0
	m.File = m.rotateFile(lname)
}

// Sync stop rountine and sync file
func (m *MLogWriter) Sync() {
	m.wait.Wait()
}

func defaultRotate(fname string) *os.File {
	f, err := os.OpenFile(fname, os.O_RDWR|os.O_APPEND|os.O_CREATE|os.O_TRUNC, 0666)
	if err != nil {
		fmt.Fprintf(os.Stderr, "OpenFile[%v] Error %v", fname, err)
		return os.Stderr
	}

	return f
}
