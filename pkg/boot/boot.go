package boot

import (
	r "crypto/rand"
	"math"
	"math/big"
	"math/rand"
	"os"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
)

var randOnce sync.Once

func InitBoot() {
	seedMathRand()
	formatter := &logrus.TextFormatter{
		ForceColors:     true,
		FullTimestamp:   true,
		TimestampFormat: "2006-01-02 15:04:05",
	}
	logrus.SetOutput(os.Stdout)
	logrus.SetFormatter(formatter)
}

func seedMathRand() {
	randOnce.Do(func() {
		n, err := r.Int(r.Reader, big.NewInt(math.MaxInt64))
		if err != nil {
			rand.Seed(time.Now().UTC().UnixNano())
		} else {
			rand.Seed(n.Int64())
		}
	})
}
