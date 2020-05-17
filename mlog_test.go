package mlog

import (
	"fmt"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"testing"
	"time"
)

func TestMLog(t *testing.T) {
	//openNewFile("aaaa")
	//ml := NewMLog()

	//os.Stderr.Write([]byte("hello\n"))
	//os.Stderr.Sync()
	//os.Stderr.Close()
	//
	//n, err := os.Stderr.Write([]byte("hello1111"))
	//if err != nil {
	//	fmt.Printf("%d - %v\n", n, err)
	//}
	//
	//return

	// output stdin
	Logger.Debug().Timestamp().Uint64("uid", 100).Msg("xxxxx....")

	Logger.Debug().Fields(map[string]interface{}{"a":1, "b":2}).Msg("ok")

	// output into file
	ml, _ := New(DebugLevel, Logfile("debug", 40), Cache(8, 127, 1, func(i int) {
		fmt.Printf("###############alert clision %v\n", i)
	}))

	time.Sleep(time.Second)
	for i:=0; i< 1000; i++ {
		go func(i int) {
			ml.Debug().Uint32("uid", uint32(i)).Msg("h78-----------------------90")
		}(i)

		if i % 100 == 0 {
			time.Sleep(time.Second)
		}
	}

	defer func() {
		ml.Close()
		s := ml.Static()
		fmt.Printf("static =%s\n", string(s))
	}()

	return

	//logger, _ := zap.NewProduction()
	//defer logger.Sync()
	//
	//logger.Info("failed to fetch URL",
	//	// Structured context as strongly typed Field values.
	//	zap.String("url", "url"),
	//	zap.Int("attempt", 3),
	//	zap.Duration("backoff", time.Second),
	//)
	//
	//sugar := logger.Sugar()
	//sugar.Infow("sugarinfo", "url", "url", "attempt", 3)
	//
	//fmt.Printf("---------------------zero log---------------------\n")
	//
	zerolog.TimeFieldFormat = ""
	//
	//w := bufio.NewWriter(nil)
	//w.Write([]byte{})
	//w.Flush()
	//
	log.Debug().
		Str("Scale", "833 cents").
		Float64("Interval", 833.09).
		Msg("Fibonacci is everywhere")
}
