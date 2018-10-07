package main

import (
	"flag"
	"fmt"
	"github.com/kristjank/ark-go/arkcoin"
	"github.com/tyler-smith/go-bip39"
	"math/rand"
	"os"
	"strings"
	"time"
)

type NetworkJob struct {
	Prefix          string
	Suffix          string
	PrefixAndSuffix bool
	CaseInsensitive bool
}

type Network struct {
	Jobs          []NetworkJob
	AddressConfig *arkcoin.Params
}

type Result struct {
	Address    string
	Passphrase string
	Matches    bool
}

var networks []Network

var entropyValue int
var addressCount int

func generate(channel chan []Result) {
	entropy, _ := bip39.NewEntropy(entropyValue)
	passphrase, _ := bip39.NewMnemonic(entropy)

	results := make([]Result, 0)
	for _, network := range networks {
		publicKey := arkcoin.NewPrivateKeyFromPassword(passphrase, network.AddressConfig).PublicKey
		address := publicKey.Address()

		for _, job := range network.Jobs {
			hasPrefix := false
			hasSuffix := false
			if len(job.Prefix) > 0 {
				if job.CaseInsensitive {
					hasPrefix = strings.HasPrefix(strings.ToLower(address), strings.ToLower(job.Prefix))
				} else {
					hasPrefix = strings.HasPrefix(address, job.Prefix)
				}
			}
			if len(job.Suffix) > 0 {
				if job.CaseInsensitive {
					hasSuffix = strings.HasSuffix(strings.ToLower(address), strings.ToLower(job.Suffix))
				} else {
					hasSuffix = strings.HasSuffix(address, job.Suffix)
				}
			}

			if (job.PrefixAndSuffix && hasPrefix && hasSuffix) || (!job.PrefixAndSuffix && (hasPrefix || hasSuffix)) {
				results = append(
					results,
					Result{
						Address:    address,
						Passphrase: passphrase,
						Matches:    true,
					},
				)
			} else {
				results = append(
					results,
					Result{
						Address:    "",
						Passphrase: "",
						Matches:    false,
					},
				)
			}
		}

		channel <- results
	}
}

func main() {
	var addressPrefix string
	var addressSuffix string
	var addressPrefixAndSuffix bool
	var addressFormat int
	var threads int
	var wif string
	var caseInsensitive bool
	var milestone int
	var fileOutput string
	flag.StringVar(&addressPrefix, "prefix", "", "Address Prefix to search for")
	flag.StringVar(&addressPrefix, "p", "", "Address Prefix to search for")
	flag.StringVar(&addressSuffix, "suffix", "", "Address Suffix to search for")
	flag.StringVar(&addressSuffix, "s", "", "Address Suffix to search for")
	flag.BoolVar(&addressPrefixAndSuffix, "prefix-and-suffix", false, "Address must include prefix and suffix")
	flag.BoolVar(&addressPrefixAndSuffix, "ps", false, "Address must include prefix and suffix")
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
	flag.IntVar(&addressCount, "count", 1, "Quantity of addresses to generate")
	flag.IntVar(&addressCount, "c", 1, "Quantity of addresses to generate")
	flag.IntVar(&milestone, "milestone", 1000000, "Milestone to log how many passphrases processed")
	flag.IntVar(&milestone, "m", 1000000, "Milestone to log how many passphrases processed")
	flag.StringVar(&fileOutput, "output", "results.txt", "File path to output results")
	flag.StringVar(&fileOutput, "o", "results.txt", "File path to output results")
	flag.Parse()

	if len(addressPrefix) <= 1 && len(addressSuffix) < 1 {
		fmt.Println("Must pass prefix and/or suffix as argument. E.g. go run vanity.go -prefix ABC -suffix DEF")

		return
	}

	if addressPrefixAndSuffix && (len(addressPrefix) == 0 || len(addressSuffix) == 0) {
		addressPrefixAndSuffix = false
	}

	if entropyValue < 128 || entropyValue > 256 {
		fmt.Println("Entropy value must be between 128 and 256")

		return
	}

	addressConfig := &arkcoin.Params{
		DumpedPrivateKeyHeader: []byte(wif),
		AddressHeader:          byte(addressFormat),
	}

	networks = append(networks, Network{
		AddressConfig: addressConfig,
		Jobs: []NetworkJob{
			{
				Prefix:          addressPrefix,
				Suffix:          addressSuffix,
				PrefixAndSuffix: addressPrefixAndSuffix,
				CaseInsensitive: caseInsensitive,
			},
		},
	})

	rand.Seed(time.Now().UTC().UnixNano())

	start := time.Now()
	channel := make(chan []Result)
	count := 0
	perBatch := threads
	batchBenchmark := false
	batchBenchmarkMax := 500
	rerunBenchmarks := false
	benchmarkCount := 0
	benchmarkRerunThreshold := 10000000
	benchmarkRun := 1
	benchmarkRunMax := 3
	if perBatch == 0 {
		perBatch = 1
		batchBenchmark = true
		rerunBenchmarks = true
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
	if len(addressPrefix) > 0 && len(addressSuffix) == 0 {
		fmt.Printf("Looking for Address with prefix '%s'\n", addressPrefix)
	} else if len(addressSuffix) > 0 && len(addressPrefix) == 0 {
		fmt.Printf("Looking for Address with suffix '%s'\n", addressSuffix)
	} else if addressPrefixAndSuffix {
		fmt.Printf("Looking for Address with prefix '%s' AND suffix '%s'\n", addressPrefix, addressSuffix)
	} else {
		fmt.Printf("Looking for Address with prefix '%s' OR suffix '%s'\n", addressPrefix, addressSuffix)
	}
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
			for _, result := range response {
				if result.Matches {
					fmt.Println("")
					fmt.Println("Address:", result.Address)
					fmt.Println("Passphrase:", result.Passphrase)

					if fileOutput != "" {
						fileHandler, err := os.OpenFile(fileOutput, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
						if err != nil {
							fmt.Println("Could not open file")
						} else {
							_, addressError := fileHandler.WriteString("Address: " + result.Address + "\n")
							_, passphraseError := fileHandler.WriteString("Passphrase: " + result.Passphrase + "\n")

							if addressError != nil || passphraseError != nil {
								fmt.Println("Could not write results to file")
							}

							fileHandler.Close()
						}
					}

					matches++
					if matches == addressCount {
						done = true
						break
					}
				}
			}
		}
		if done {
			break
		}
		if rerunBenchmarks && !batchBenchmark && benchmarkCount >= benchmarkRerunThreshold {
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
					benchmarkCount = 0
					perBatch = bestBatch
					fmt.Println("")
					fmt.Println("Batch", perBatch, "processed", int(bestPms), "per ms")
					fmt.Println("Benchmark complete. Threads set to", perBatch)
				} else {
					perBatch = 1
					benchmarkRun++
				}
			}
		}
	}

	elapsedSoFar := time.Now().Sub(start)
	fmt.Println("Checked", count, "passphrases within", elapsedSoFar)
}
