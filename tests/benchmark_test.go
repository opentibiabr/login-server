package tests

import (
	"bytes"
	"log"
	"net/http"
	"sync"
	"testing"
	"time"
)

var wg sync.WaitGroup

var countOk = 0

func asynRequest(payload []byte) {
	res, _ := http.Post("http://127.0.0.1:80/login2", "application/json", bytes.NewBuffer(payload))

	if res != nil && res.StatusCode == http.StatusOK {
		countOk++
	}

	wg.Done()
}

func BenchmarkApi(b *testing.B) {
	if b.N <= 1 {
		return
	}
	payload := []byte(`{"type":"login","email":"@god","password":"god"}`)
	log.Print("\nBenchmarking with OpenTibiaBR Login Server")

	totalTime := int64(0)
	for j := 0; j < 10; j++ {
		wg.Add(b.N)
		start := time.Now()
		for i := 0; i < b.N; i++ {
			go asynRequest(payload)
		}
		wg.Wait()
		requestTime := time.Since(start).Milliseconds()
		log.Printf("performing %dx requests in %dms", b.N, requestTime)
		totalTime += requestTime
		time.Sleep(100 * time.Millisecond)
	}
	log.Printf("average: %.2f requests/s", float64(10*b.N)/float64(totalTime)*1000)
	log.Printf("availability: %.4f", float64(countOk)/float64(10*b.N))
}
