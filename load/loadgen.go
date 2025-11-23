package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"net"
	"net/http"
	"strconv"
	"sync"
	"sync/atomic"
	"time"
)

var (
	serverURL  string
	threads    int
	duration   int
	startKey   int64
	reqTimeout int
)

type KVRequest struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

func newHTTPClient(timeout time.Duration) *http.Client {
	transport := &http.Transport{
		MaxIdleConns:        1000,
		MaxIdleConnsPerHost: 1000,
		MaxConnsPerHost:     0, // unlimited, OS limits apply
		DialContext: (&net.Dialer{
			Timeout:   5 * time.Second,
			KeepAlive: 30 * time.Second,
		}).DialContext,
	}
	return &http.Client{
		Timeout:   timeout,
		Transport: transport,
	}
}

func worker(
	id int,
	wg *sync.WaitGroup,
	successCount *int64,
	errorCount *int64,
	totalLatencyMicros *int64,
	endTime time.Time,
	client *http.Client,
) {
	defer wg.Done()

	for {
		if time.Now().After(endTime) {
			return
		}

		// Generate unique numeric key
		key := atomic.AddInt64(&startKey, 1)
		reqData := KVRequest{
			Key:   strconv.FormatInt(key, 10),
			Value: "DB_Load_Test",
		}

		reqJSON, err := json.Marshal(reqData)
		if err != nil {
			atomic.AddInt64(errorCount, 1)
			continue
		}

		req, err := http.NewRequest("POST", serverURL+"/api/kv", bytes.NewBuffer(reqJSON))
		if err != nil {
			atomic.AddInt64(errorCount, 1)
			continue
		}
		req.Header.Set("Content-Type", "application/json")

		start := time.Now()
		resp, err := client.Do(req)
		latencyMicros := time.Since(start).Microseconds()

		if err != nil {
			atomic.AddInt64(errorCount, 1)
			continue
		}

		// Always close body to reuse connection
		_ = resp.Body.Close()

		if resp.StatusCode == http.StatusOK || resp.StatusCode == http.StatusCreated {
			atomic.AddInt64(successCount, 1)
			atomic.AddInt64(totalLatencyMicros, latencyMicros)
		} else {
			atomic.AddInt64(errorCount, 1)
		}

		// OPTIONAL: Think time per client (commented out for max load)
		// time.Sleep(1 * time.Millisecond)
	}
}

func main() {
	flag.StringVar(&serverURL, "server", "http://localhost:8080", "Server base URL")
	flag.IntVar(&threads, "threads", 10, "Number of concurrent clients (goroutines)")
	flag.IntVar(&duration, "duration", 30, "Test duration (seconds)")
	flag.Int64Var(&startKey, "startKey", 1, "Initial numeric key")
	flag.IntVar(&reqTimeout, "timeout", 2000, "Per-request timeout in milliseconds")
	flag.Parse()

	fmt.Println("Starting PutAll workload...")
	fmt.Printf("Server: %s | Threads: %d | Duration: %ds | StartKey: %d\n",
		serverURL, threads, duration, startKey)

	var wg sync.WaitGroup
	var successCount int64
	var errorCount int64
	var totalLatencyMicros int64

	endTime := time.Now().Add(time.Duration(duration) * time.Second)

	// One HTTP client per worker to avoid lock contention on a single client
	clientTimeout := time.Duration(reqTimeout) * time.Millisecond

	wg.Add(threads)
	for i := 0; i < threads; i++ {
		client := newHTTPClient(clientTimeout)
		go worker(i, &wg, &successCount, &errorCount, &totalLatencyMicros, endTime, client)
	}

	wg.Wait()

	totalReq := successCount + errorCount
	testTimeSeconds := float64(duration)

	var throughput float64
	if testTimeSeconds > 0 {
		throughput = float64(successCount) / testTimeSeconds
	}

	var avgLatencyMs float64
	if successCount > 0 {
		avgLatencyMs = float64(totalLatencyMicros) / float64(successCount) / 1000.0
	}

	var errorPercentage float64
	if totalReq > 0 {
		errorPercentage = (float64(errorCount) / float64(totalReq)) * 100.0
	}

	fmt.Println("----------- Load Test Results -----------")
	fmt.Printf("Total Requests Sent   : %d\n", totalReq)
	fmt.Printf("Success Count         : %d\n", successCount)
	fmt.Printf("Error Count           : %d\n", errorCount)
	fmt.Printf("Error Percentage      : %.2f %%\n", errorPercentage)
	fmt.Printf("Throughput            : %.2f req/s\n", throughput)
	fmt.Printf("Average Response Time : %.3f ms\n", avgLatencyMs)
	fmt.Println("----------------------------------------")
}
