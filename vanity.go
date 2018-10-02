package main

import (
	"flag"
	"fmt"
	"github.com/kristjank/ark-go/arkcoin"
	"github.com/tyler-smith/go-bip39"
	"math/rand"
	"strconv"
	"strings"
	"time"
)

var flag_prefix string
var flag_entropy string
var flag_publicKeyHash string
var flag_threads string
var flag_wif string

var entropyBase int
var addressConfig = arkcoin.ArkCoinMain

func generate(channel chan []string) {
	entropy, _ := bip39.NewEntropy(entropyBase)
	passphrase, _ := bip39.NewMnemonic(entropy)
	publicKey := arkcoin.NewPrivateKeyFromPassword(passphrase, addressConfig).PublicKey
	address := publicKey.Address()
	if strings.Index(address, flag_prefix) == 0 {
		channel <- []string{passphrase, address, "Y"}
	} else {
		channel <- []string{passphrase, address, ""}
	}
}

func main() {
	flag.StringVar(&flag_prefix, "prefix", "", "Specify entropy")
	flag.StringVar(&flag_entropy, "entropy", "128", "Specify entropy")
	flag.StringVar(&flag_publicKeyHash, "public-key-hash", "23", "Address Prefix")
	flag.StringVar(&flag_threads, "threads", "100", "Threads to run")
	flag.StringVar(&flag_wif, "wif", "170", "WIF")
	flag.Parse()

	if len(flag_prefix) <= 1 {
		fmt.Println("Must pass prefix as argument. E.g. go run vanity.go -prefix ABCDEFG")
		return
	}

	entropyValue, err := strconv.Atoi(flag_entropy)
	if err != nil {
		fmt.Println("There was a problem parsing the entropy argument")
		return
	}
	if entropyValue < 128 || entropyValue > 256 {
		fmt.Println("Entropy value must be between 128 and 256")
		return
	}
	entropyBase = entropyValue

	publicKeyHashValue, err := strconv.Atoi(flag_publicKeyHash)
	if err != nil {
		fmt.Println("There was a problem parsing the address prefix argument")
		return
	}

	addressConfig = &arkcoin.Params{
		DumpedPrivateKeyHeader: []byte(flag_wif),
		AddressHeader:          byte(publicKeyHashValue),
	}

	rand.Seed(time.Now().UTC().UnixNano())

	start := time.Now()
	channel := make(chan []string)
	count := 0

	threadsValue, err := strconv.Atoi(flag_threads)
	if err != nil {
		fmt.Println("There was a problem parsing the threads argument")
		return
	}

	perBatch := threadsValue
	done := false
	fmt.Println("Looking for Address with prefix:", flag_prefix)
	fmt.Println("")
	for {
		for i := 0; i < perBatch; i++ {
			go generate(channel)
		}
		for i := 0; i < perBatch; i++ {
			count++
			if (count % 100000) == 0 {
				elapsedSoFar := time.Now().Sub(start)
				fmt.Println("Checked", count, "passphrases within", elapsedSoFar)
			}
			response := <-channel
			if response[2] == "Y" {
				fmt.Println("Passphrase:", response[0])
				fmt.Println("Address:", response[1])
				done = true
				break
			}
		}
		if done {
			break
		}
	}
}
