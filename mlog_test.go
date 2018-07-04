package mlog

import (
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	//"go.uber.org/zap"
	"testing"
	"fmt"
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

	ml := NewMLog(DebugLevel, FileName("test", 40))
	time.Sleep(time.Second)
	ml.Debug().Uint32("uid", 1023).Msg("h78-----------------------90")
	time.Sleep(time.Second)
	ml.Debug().Uint32("uid", 1023).Msg("hehehe111111111111111111111")
	time.Sleep(time.Second)
	ml.Debug().Uint32("uid", 1023).Msg("hehehe1234")
	time.Sleep(time.Second)

	defer func() {
		ml.Sync()
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
