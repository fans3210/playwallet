package tools

import (
	"fmt"
	"log"
	"math"
	"math/rand"
	"time"
)

func Retry(maxRetry int, f func() error) error {
	var err error
	for i := 0; i < maxRetry; i++ {
		if e := f(); e != nil {
			err = e
			log.Printf("%dth attempt failed due to err: %s, prepare to retry, maxRety = %d\n", i+1, err, maxRetry)
			t := float64(100 + rand.Int63n(1000))

			time.Sleep(time.Duration(t*math.Pow(2, float64(i+1))) * time.Millisecond)
			continue
		}
		return nil
	}
	return fmt.Errorf("reached maxRetry %d. last err: %w", maxRetry, err)
}
