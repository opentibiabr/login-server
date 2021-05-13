package tests

import (
	"bytes"
	"net/http"
	"sync"
	"testing"
)

var wg sync.WaitGroup

func asynRequest(payload []byte) {
	_,err := http.Post("http://localhost:80/login", "application/json", bytes.NewBuffer(payload))
	if err != nil {}
	wg.Done()
}

func BenchmarkApi(b *testing.B) {
	wg.Add(b.N * 1)
	b.ResetTimer()
	b.StopTimer()
	payload := []byte(`{"type":"login","email":"@god","password":"god"}`)
	for i := 0; i < b.N; i++ {
		b.StartTimer()
		go asynRequest(payload)
		b.StopTimer()
	}
	wg.Wait()
}
