package tests

import (
	"bytes"
	"context"
	"errors"
	"github.com/opentibiabr/login-server/src/grpc/login_proto_messages"
	"github.com/opentibiabr/login-server/src/logger"
	"google.golang.org/grpc"
	"log"
	"net/http"
	"sync"
	"testing"
	"time"
)

var wg sync.WaitGroup

var countOk = 0

func asynRequest(payload []byte) {
	res, _ := http.Post("http://localhost:80/login", "application/json", bytes.NewBuffer(payload))

	if res != nil && res.StatusCode == http.StatusOK {
		countOk++
	}

	wg.Done()
}

func BenchmarkApi(b *testing.B) {
	countOk = 0

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

func asyncGrpcCall(conn *grpc.ClientConn, payload *login_proto_messages.LoginRequest) {
	grpcClient := login_proto_messages.NewLoginServiceClient(conn)
	res, err := grpcClient.Login(
		context.Background(),
		payload,
	)

	if err == nil && res.GetError() == nil {
		countOk++
	}

	wg.Done()
}

func BenchmarkTcp(b *testing.B) {
	countOk = 0
	if b.N <= 1 {
		return
	}

	conn, err := grpc.Dial(":7171", grpc.WithInsecure())
	if err != nil {
		logger.Error(errors.New("Couldn't start GRPC reverse proxy."))
	}

	payload := login_proto_messages.LoginRequest{
		Email:    "@god",
		Password: "god",
	}

	log.Print("\nBenchmarking with gRPC OpenTibiaBR Login Server")

	totalTime := int64(0)
	for j := 0; j < 10; j++ {
		wg.Add(b.N)
		start := time.Now()
		for i := 0; i < b.N; i++ {
			go asyncGrpcCall(conn, &payload)
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
