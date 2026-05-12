package main

import (
	"flag"
	"fmt"
	"time"

	"github.com/joho/godotenv"
)

var (
	target    = flag.String("target", "http://localhost:8002", "ingest server base URL")
	apiKey    = flag.String("api-key", "", "API key for ingest auth header")
	hmacSec   = flag.String("hmac-secret", "", "HMAC secret for request signing")
	svc       = flag.String("service", "checkout-svc", "service name to emit logs for")
	rps       = flag.Int("rps", 10, "requests per second")
	errorRate = flag.Float64("error-rate", 0.05, "fraction of requests that emit error logs (0.0–1.0)")
	duration  = flag.Duration("duration", 60*time.Second, "how long to run the simulation")
)

func main() {
	flag.Parse()
	_ = godotenv.Load()

	fmt.Printf("Pulse Sim starting: target=%s svc=%s rps=%d err=%.0f%% duration=%s\n",
		*target, *svc, *rps, *errorRate*100, *duration)

	// TODO TASK-024: implement HTTP POST batch to POST /v1/ingest/logs
	// batch log entries at *rps per second, inject errors at *errorRate,
	// sign each request with HMAC-SHA256 using *hmacSec, auth via *apiKey header.
	// Run for *duration then exit 0.

	fmt.Println("Pulse Sim: simulation complete (stub — implement TASK-024)")
}
