package main

import (
	"fmt"
	"time"
)

type BenchmarkConfig struct {
	Enabled        bool
	ThreadMax      int
	Count          int
	Rerun          bool
	RerunThreshold int
	Run            int
	RunMax         int
	Batches        map[int][]BenchmarkResult
}

type BenchmarkResult struct {
	start    time.Time
	duration time.Duration
	count    float64
	perMs    float64
}

var benchmark *BenchmarkConfig

func ProcessBenchmark(batchResult *BenchmarkResult) {
	if config.Threads < benchmark.ThreadMax {
		batchResult.duration = time.Now().Sub(batchResult.start)
		batchResult.perMs = batchResult.count / (batchResult.duration.Seconds() * 1000)
		benchmark.Batches[config.Threads] = append(benchmark.Batches[config.Threads], *batchResult)
		config.Threads++
	} else {
		if benchmark.Run >= benchmark.RunMax {
			bestBatch := 1
			var bestPms float64
			for i := 1; i <= benchmark.ThreadMax; i++ {
				var totalPms float64
				for p := 0; p < len(benchmark.Batches[i]); p++ {
					totalPms += benchmark.Batches[i][p].perMs
				}
				pmsAverage := totalPms / float64(len(benchmark.Batches[i]))
				if pmsAverage > bestPms {
					bestBatch = i
					bestPms = pmsAverage
				}
			}
			benchmark.Enabled = false
			benchmark.Count = 0
			config.Threads = bestBatch
			fmt.Println("")
			fmt.Println("Batch", config.Threads, "processed", int(bestPms), "per ms")
			fmt.Println("Benchmark complete. Threads set to", config.Threads)
		} else {
			config.Threads = 1
			benchmark.Run++
		}
	}
}
