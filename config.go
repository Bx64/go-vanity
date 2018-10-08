package main

import (
	"flag"
	"log"

	"github.com/kristjank/ark-go/arkcoin"
)

type GlobalConfig struct {
	Entropy         int
	Count           int
	Threads         int
	FileOutput      string
	Networks        []Network
	RunIndefinitely bool
	LoadedConfig    bool
}

type NetworkJob struct {
	Prefix          string
	Suffix          string
	PrefixAndSuffix bool
	CaseInsensitive bool
	Skip            bool
}

type Network struct {
	Jobs          []NetworkJob
	AddressConfig *arkcoin.Params
	Wif           int
	AddressFormat int
}

type Result struct {
	Address    string
	Passphrase string
	Matches    bool
}

func (job *NetworkJob) IsValid() bool {
	if len(job.Prefix) <= 1 && len(job.Suffix) < 1 {
		return false
	}

	if job.PrefixAndSuffix && (len(job.Prefix) == 0 || len(job.Suffix) == 0) {
		job.PrefixAndSuffix = false
	}

	return true
}

var config = &GlobalConfig{}

func LoadConfig() {
	var addressPrefix string
	var addressSuffix string
	var addressPrefixAndSuffix bool
	var addressFormat int
	var wif int
	var caseInsensitive bool

	flag.StringVar(&addressPrefix, "prefix", "", "Address Prefix to search for")
	flag.StringVar(&addressPrefix, "p", "", "Address Prefix to search for")
	flag.StringVar(&addressSuffix, "suffix", "", "Address Suffix to search for")
	flag.StringVar(&addressSuffix, "s", "", "Address Suffix to search for")
	flag.BoolVar(&addressPrefixAndSuffix, "prefix-and-suffix", false, "Address must include prefix and suffix")
	flag.BoolVar(&addressPrefixAndSuffix, "ps", false, "Address must include prefix and suffix")
	flag.IntVar(&config.Entropy, "entropy", 128, "Specify entropy")
	flag.IntVar(&config.Entropy, "e", 128, "Specify entropy")
	flag.IntVar(&addressFormat, "address-format", 23, "Address Format")
	flag.IntVar(&addressFormat, "a", 23, "Address Format")
	flag.IntVar(&config.Threads, "threads", 100, "Threads to run")
	flag.IntVar(&config.Threads, "t", 100, "Threads to run")
	flag.IntVar(&wif, "wif", 170, "WIF")
	flag.IntVar(&wif, "w", 170, "WIF")
	flag.BoolVar(&caseInsensitive, "case-insensitive", false, "Case insensitive")
	flag.BoolVar(&caseInsensitive, "i", false, "Case insensitive")
	flag.IntVar(&config.Count, "count", 1, "Quantity of addresses to generate")
	flag.IntVar(&config.Count, "c", 1, "Quantity of addresses to generate")
	flag.StringVar(&config.FileOutput, "output", "results.txt", "File path to output results")
	flag.StringVar(&config.FileOutput, "o", "results.txt", "File path to output results")
	flag.Parse()

	config.Networks = []Network{
		{
			Wif:           wif,
			AddressFormat: addressFormat,
			Jobs: []NetworkJob{
				{
					Prefix:          addressPrefix,
					Suffix:          addressSuffix,
					PrefixAndSuffix: addressPrefixAndSuffix,
					CaseInsensitive: caseInsensitive,
				},
			},
		},
	}

	if config.Entropy < 128 || config.Entropy > 256 {
		log.Fatalln("Entropy value must be between 128 and 256")
	}

	for networkId := range config.Networks {
		network := &config.Networks[networkId]
		network.AddressConfig = &arkcoin.Params{
			DumpedPrivateKeyHeader: []byte(string(network.Wif)),
			AddressHeader:          byte(network.AddressFormat),
		}
		for jobId := range network.Jobs {
			job := &network.Jobs[jobId]
			if !job.IsValid() {
				job.Skip = true
			}
		}
	}
}
