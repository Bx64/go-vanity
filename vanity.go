package main

import (
	"fmt"
	"math/rand"
	"os"
	"strings"
	"time"

	"github.com/kristjank/ark-go/arkcoin"
	"github.com/tyler-smith/go-bip39"
)

func generate(channel chan []Result) {
	entropy, _ := bip39.NewEntropy(config.Entropy)
	passphrase, _ := bip39.NewMnemonic(entropy)

	results := make([]Result, 0)
	for _, network := range config.Networks {
		publicKey := arkcoin.NewPrivateKeyFromPassword(passphrase, network.AddressConfig).PublicKey
		address := publicKey.Address()

		for _, job := range network.Jobs {
			if job.Skip {
				continue
			}

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
				alreadyFound := false
				for _, result := range results {
					if result.Address == address {
						alreadyFound = true
						break
					}
				}
				if alreadyFound {
					continue
				}

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
	LoadConfig()

	rand.Seed(time.Now().UTC().UnixNano())

	start := time.Now()
	channel := make(chan []Result)
	matchCount := 0

	benchmark = &BenchmarkConfig{
		ThreadMax:      500,
		Rerun:          false,
		RerunThreshold: 10000000,
		Run:            1,
		RunMax:         3,
		Batches:        make(map[int][]BenchmarkResult, 500),
	}

	if config.Threads == 0 {
		config.Threads = 1
		benchmark.Enabled = true
		benchmark.Rerun = true
	}

	config.RunIndefinitely = !benchmark.Enabled
	searchFinished := false
	if benchmark.Enabled {
		fmt.Println("Benchmarking...")
	}
	jobCount := 0
	for _, network := range config.Networks {
		jobCount += len(network.Jobs)
	}
	if config.LoadedConfig {
		fmt.Printf("Loaded config file with %v networks totalling %v jobs\n\n", len(config.Networks), jobCount)

		fmt.Println("Only the following options are used when using a file import:")
		fmt.Println("  entropy, threads, output")
	} else {
		addressPrefix := config.Networks[0].Jobs[0].Prefix
		addressSuffix := config.Networks[0].Jobs[0].Suffix
		addressPrefixAndSuffix := config.Networks[0].Jobs[0].PrefixAndSuffix
		if len(addressPrefix) > 0 && len(addressSuffix) == 0 {
			fmt.Printf("Looking for Address with prefix '%s'\n", addressPrefix)
		} else if len(addressSuffix) > 0 && len(addressPrefix) == 0 {
			fmt.Printf("Looking for Address with suffix '%s'\n", addressSuffix)
		} else if addressPrefixAndSuffix {
			fmt.Printf("Looking for Address with prefix '%s' AND suffix '%s'\n", addressPrefix, addressSuffix)
		} else {
			fmt.Printf("Looking for Address with prefix '%s' OR suffix '%s'\n", addressPrefix, addressSuffix)
		}
	}
	for {
		benchmarkResult := &BenchmarkResult{
			start:    time.Now(),
			duration: 0,
			count:    0,
		}
		for i := 0; i < config.Threads; i++ {
			go generate(channel)
		}
		for i := 0; i < config.Threads || config.RunIndefinitely; i++ {
			matchCount++
			benchmark.Count++
			benchmarkResult.count++
			if benchmark.Count >= benchmark.RerunThreshold {
				config.RunIndefinitely = false
			}
			elapsedSoFar := time.Now().Sub(start)
			fmt.Printf("\033[2KChecked %v passphrases within %v\r", matchCount, elapsedSoFar)
			response := <-channel
			for _, result := range response {
				if result.Matches {
					fmt.Println(strings.Repeat(" ", 100))
					fmt.Println("Address:", result.Address)
					fmt.Println("Passphrase:", result.Passphrase)

					if config.FileOutput != "" {
						fileHandler, err := os.OpenFile(config.FileOutput, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
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

					matchCount++
					if config.Count != 0 && matchCount == config.Count {
						searchFinished = true
						break
					}
				}
			}
			if config.RunIndefinitely {
				go generate(channel)
			}
		}
		if searchFinished {
			break
		}
		if benchmark.Rerun && !benchmark.Enabled && benchmark.Count >= benchmark.RerunThreshold {
			config.Threads = 1
			benchmark.Enabled = true
			benchmark.Run = 1
			benchmark.Batches = make(map[int][]BenchmarkResult, benchmark.ThreadMax)
			if benchmark.Enabled {
				fmt.Println("")
				fmt.Println("Benchmarking...")
			}
		}
		if benchmark.Enabled {
			ProcessBenchmark(benchmarkResult)
		}
	}

	elapsedSoFar := time.Now().Sub(start)
	fmt.Println("Checked", matchCount, "passphrases within", elapsedSoFar)
}
