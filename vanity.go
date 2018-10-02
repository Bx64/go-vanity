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
var addressConfig = arkcoin.ArkCoinMain

func generate(channel chan []string) {
	entropy, _ := bip39.NewEntropy(entropyValue)
	passphrase, _ := bip39.NewMnemonic(entropy)
	publicKey := arkcoin.NewPrivateKeyFromPassword(passphrase, addressConfig).PublicKey
	address := publicKey.Address()
	if strings.Index(address, addressPrefix) == 0 {
		channel <- []string{passphrase, address, "Y"}
	} else {
		channel <- []string{passphrase, address, ""}
	}
}

func main() {
	var addressFormat int
	var threads int
	var wif string
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
	done := false
	fmt.Println("Looking for Address with prefix:", addressPrefix)
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
