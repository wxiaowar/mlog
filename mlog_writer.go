package mlog

import (
	"encoding/json"
	"fmt"
	"os"
	"sync"
	"sync/atomic"
	"time"
)

const (
	NoSync = 0
	IsSync = 1

	block = 1 << 13  // 8k
	k64 = 1 << 16 // 64KiB
)

var pool = &sync.Pool{
	New: func() interface{} {
		buf := make([]byte, 0, 512)
		return &RingItem{Data:buf}
	},
}

// MWriter write log to file
type writer struct {
	*os.File

	name string // log file name
	tick int64  // tick time interval
	max  int64  // log file max size
	cur  int64  // cur size

	cache *Ring // cache for performance
	cacheMax int
	cacheCur int

	status  int32 // sync status

	exit     bool
	exitChan chan bool
}

var defaultOut = os.Stderr

// NewWriter create writer
func newWriter(options ...Option) (*writer, error) {
	ml := &writer{
		max:      int64(1 << 61),
		cache:    nil,
		File:     defaultOut,
		exitChan: make(chan bool),
	}

	for _, opt := range options {
		opt.apply(ml)
	}

	if ml.name == "" {
		return ml, nil
	}

	err := ml.open()
	if err != nil {
		return nil, err
	}

	ml.cur = ml.logFileSize()

	if ml.cache != nil {
		go ml.timeTick()
	}

	return ml, err
}

// Static log info
func (m *writer) Static() []byte {
	ret := map[string]interface{}{
		"cnt":  m.cacheCur,
		"max": m.cacheMax,
	}

	bts, err := json.Marshal(ret)
	if err != nil {
		fmt.Fprintf(os.Stderr, "MLog Static Error %v\n", err)
	}

	return bts
}

// Write Note: should stop logic before stop log
func (m *writer) Write(p []byte) (n int, err error) {
	if m.cache == nil { // dir output into disk
		return m.File.Write(p)
	}

	// write to cache
	m.cacheCur++
	item := pool.Get().(*RingItem)
	item.Data = append(item.Data, p...)
	m.cache.Add(item)

	// trigger write disk
	//fmt.Printf("write %v-%v-%v\n", m.cacheCur, m.cacheMax, m.status)
	if m.cacheCur > m.cacheMax && m.status == NoSync &&
		atomic.CompareAndSwapInt32(&m.status, NoSync, IsSync) {
		go m.flush()
	}

	return len(p), nil
}

// close sync file
func (m *writer) Close() {
	if m.exit {
		return
	}

	m.exit = true
	<-m.exitChan
}

func (m *writer) timeTick() {
	if m.cache == nil {
		return
	}

	ticker := time.NewTicker(time.Duration(m.tick) * time.Second)
	for range ticker.C {
		if atomic.CompareAndSwapInt32(&m.status, NoSync, IsSync) {
			m.flush()
		}

		if m.exit && m.cacheCur == 0 {
			m.exitChan <- true
			return
		}
	}
}

//
func (m *writer) flush() {
	defer func() {
		m.status = NoSync
	}()

	if m.cache == nil {
		return
	}

	size := 0
	ext := []byte(nil)
	buff := make([]byte, 0, block)
	for {
		item, ok := m.cache.Next()
		if !ok {
			m.cacheCur = 0
			break
		}

		if item.Data == nil {
			continue
		}

		need := len(item.Data)
		if size + need > block {
			ext = make([]byte, 0, need)
			copy(ext, item.Data)

			m.cacheCur = m.cacheMax + 1 // for next trigger
			break  // ignore this, not put into pool
		}

		buff = append(buff, item.Data...)
		size += need

		if cap(item.Data) <= k64 {
			item.Data = item.Data[:0]
			pool.Put(item)
		}
	}

	if size < 1 {
		return
	}

	// read data form cache
	n, err := m.File.Write(buff)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Write [%s] Error %v\n", string(buff), err)
	} else {
		m.cur += int64(n)
	}

	// ext
	if ext != nil {
		n, err := m.File.Write(ext)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Write [%s] Error %v\n", string(ext), err)
		} else {
			m.cur += int64(n)
		}
	}

	//fmt.Printf("buff: %v-%v-%v-%v-%s\n", n, err, m.cur, m.max, string(buff))
	// sync data to disk
	m.File.Sync()

	//
	if m.cur > m.max { // rotate
		m.reopen()
		m.cur = m.logFileSize()
		//fmt.Printf("reopen.....%d-%d\n", m.cur, m.max)
	}
}

func (m writer) logName() string {
	return m.name + ".log"
}

func (m writer) archiveName() string {
	return fmt.Sprintf("%s_%s.log", m.name, time.Now().Format("2006-01-02_15-04-05"))
}

func (m *writer) reopen() error {
	if m.File == defaultOut {
		return nil
	}

	m.close()

	name := m.logName()
	arName := m.archiveName()
	//fmt.Printf("%v-%v-----\n", name, arName)
	err := os.Rename(name, arName)
	if err != nil {
		fmt.Fprintf(os.Stderr, "rename file[%v=>%v] err %v\n", name, arName, err)
	}

	return m.open()
}

func (m *writer) close() {
	m.File.Sync()
	m.File.Close()
}

func (m *writer) logFileSize() int64 {
	info, err := m.File.Stat()
	if err != nil {
		return 0
	}

	return info.Size()
}

func (m *writer) open() error {
	f, err := os.OpenFile(m.logName(), os.O_RDWR|os.O_APPEND|os.O_CREATE|os.O_TRUNC, 0666)
	if err != nil {
		fmt.Fprintf(os.Stderr, "log open file [%v] rrror %v", m.logName(), err)
		m.File = defaultOut
		return err
	}

	m.File = f
	return nil
}
