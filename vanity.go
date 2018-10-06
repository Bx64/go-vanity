package main

import (
	"flag"
	"fmt"
	"github.com/kristjank/ark-go/arkcoin"
	"github.com/tyler-smith/go-bip39"
	"math/rand"
	"strings"
	"time"
)

var addressPrefix string
var entropyValue int
var addressMax int
var caseInsensitive = false
var addressConfig = arkcoin.ArkCoinMain

func generate(channel chan []string) {
	entropy, _ := bip39.NewEntropy(entropyValue)
	passphrase, _ := bip39.NewMnemonic(entropy)
	publicKey := arkcoin.NewPrivateKeyFromPassword(passphrase, addressConfig).PublicKey
	address := publicKey.Address()
	if caseInsensitive && strings.Index(strings.ToLower(address), strings.ToLower(addressPrefix)) == 0 {
		channel <- []string{passphrase, address, "Y"}
	} else if strings.Index(address, addressPrefix) == 0 {
		channel <- []string{passphrase, address, "Y"}
	} else {
		channel <- []string{"", "", ""}
	}
}

func main() {
	var addressFormat int
	var threads int
	var wif string
	var milestone int
	flag.StringVar(&addressPrefix, "prefix", "", "Specify entropy")
	flag.StringVar(&addressPrefix, "p", "", "Specify entropy")
	flag.IntVar(&entropyValue, "entropy", 128, "Specify entropy")
	flag.IntVar(&entropyValue, "e", 128, "Specify entropy")
	flag.IntVar(&addressFormat, "address-format", 23, "Address Format")
	flag.IntVar(&addressFormat, "a", 23, "Address Format")
	flag.IntVar(&threads, "threads", 100, "Threads to run")
	flag.IntVar(&threads, "t", 100, "Threads to run")
	flag.StringVar(&wif, "wif", "170", "WIF")
	flag.StringVar(&wif, "w", "170", "WIF")
	flag.BoolVar(&caseInsensitive, "case-insensitive", false, "Case insensitive")
	flag.BoolVar(&caseInsensitive, "i", false, "Case insensitive")
	flag.IntVar(&addressMax, "count", 1, "Quantity of addresses to generate")
	flag.IntVar(&addressMax, "c", 1, "Quantity of addresses to generate")
	flag.IntVar(&milestone, "milestone", 1000000, "Milestone to log how many passphrases processed")
	flag.IntVar(&milestone, "m", 1000000, "Milestone to log how many passphrases processed")
	flag.Parse()

	if len(addressPrefix) <= 1 {
		fmt.Println("Must pass prefix as argument. E.g. go run vanity.go -prefix ABCDEFG")
		return
	}

	if entropyValue < 128 || entropyValue > 256 {
		fmt.Println("Entropy value must be between 128 and 256")
		return
	}

	addressConfig = &arkcoin.Params{
		DumpedPrivateKeyHeader: []byte(wif),
		AddressHeader:          byte(addressFormat),
	}

	rand.Seed(time.Now().UTC().UnixNano())

	start := time.Now()
	channel := make(chan []string)
	count := 0
	perBatch := threads
	batchBenchmark := false
	batchBenchmarkMax := 500
	benchmarkCount := 0
	benchmarkRerunThreshold := 10000000
	benchmarkRun := 1
	benchmarkRunMax := 5
	if perBatch == 0 {
		perBatch = 1
		batchBenchmark = true
	}
	done := false
	matches := 0
	type benchmarkResult struct {
		start    time.Time
		duration time.Duration
		count    float64
		perMs    float64
	}
	batches := make(map[int][]benchmarkResult, batchBenchmarkMax)
	if batchBenchmark {
		fmt.Println("Benchmarking...")
	}
	fmt.Println("Looking for Address with prefix:", addressPrefix)
	fmt.Println("")
	for {
		batchResult := benchmarkResult{
			start:    time.Now(),
			duration: 0,
			count:    0,
		}
		for i := 0; i < perBatch; i++ {
			go generate(channel)
		}
		for i := 0; i < perBatch; i++ {
			count++
			benchmarkCount++
			batchResult.count++
			if (count % milestone) == 0 {
				elapsedSoFar := time.Now().Sub(start)
				fmt.Printf("\033[2KChecked %d passphrases within %s\r", count, elapsedSoFar)
			}
			response := <-channel
			if response[2] == "Y" {
				fmt.Println("Address:", response[1])
				fmt.Println("Passphrase:", response[0])
				matches++
				if matches == addressMax {
					done = true
					break
				}
			}
		}
		if done {
			break
		}
		if benchmarkCount >= benchmarkRerunThreshold {
			batchBenchmark = true
			perBatch = 1
			batches = make(map[int][]benchmarkResult, batchBenchmarkMax)
			if batchBenchmark {
				fmt.Println("")
				fmt.Println("Benchmarking...")
			}
		}
		if batchBenchmark {
			if perBatch < batchBenchmarkMax {
				batchResult.duration = time.Now().Sub(batchResult.start)
				batchResult.perMs = batchResult.count / (batchResult.duration.Seconds() * 1000)
				batches[perBatch] = append(batches[perBatch], batchResult)
				perBatch++
			} else {
				if benchmarkRun >= benchmarkRunMax {
					bestBatch := 1
					var bestPms float64
					for i := 1; i <= batchBenchmarkMax; i++ {
						var totalPms float64
						for p := 0; p < len(batches[i]); p++ {
							totalPms += batches[i][p].perMs
						}
						pmsAverage := totalPms / float64(len(batches[i]))
						if pmsAverage > bestPms {
							bestBatch = i
							bestPms = pmsAverage
						}
					}
					batchBenchmark = false
					perBatch = bestBatch
					fmt.Println("")
					fmt.Println("Batch", perBatch, "processed", int(bestPms), "per ms")
					fmt.Println("Benchmark complete. Threads set to", perBatch)
				}
				benchmarkRun++
			}
		}
	}

	elapsedSoFar := time.Now().Sub(start)
	fmt.Println("Checked", count, "passphrases within", elapsedSoFar)
}
