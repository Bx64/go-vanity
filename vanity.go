package main

import (
    "fmt"
    "math/rand"
    "strings"
    "strconv"
    "time"
    "os"
    "github.com/kristjank/ark-go/arkcoin"
    "github.com/tyler-smith/go-bip39"
)

func generate(channel chan []string, prefix string) {
  var passphraseArray [12]string
  wordCount := len(bip39.WordList)
  for i := 0; i < len(passphraseArray); i++ {
    randNum := rand.Intn(wordCount)
    passphraseArray[i] = bip39.WordList[randNum]
  }
  passphrase := strings.Join(passphraseArray[:], " ")
  publicKey := arkcoin.NewPrivateKeyFromPassword(passphrase, arkcoin.ArkCoinMain).PublicKey
  address := publicKey.Address()
  if strings.Index(address, prefix) == 0 {
    channel <- []string{passphrase, address, "Y"}
  } else {
    channel <- []string{passphrase, address, ""}
  }
}

func main() {
  if (len(os.Args) == 1) {
    fmt.Println("Must pass prefix as argument. E.g. go run vanity.go ABCDEFG")
    return
  }

  rand.Seed(time.Now().UTC().UnixNano())

  start := time.Now()
  prefix := os.Args[1]
  channel := make(chan []string)
  count := 0
  perBatch := 100
  if (len(os.Args) >= 3) {
    perBatch, _ = strconv.Atoi(os.Args[2])
  }
  done := false
  fmt.Println("Looking for Address with prefix:", prefix)
  fmt.Println("")
  for {
    for i := 0; i < perBatch; i++ {
      go generate(channel, prefix)
    }
    for i := 0; i < perBatch; i++ {
      count++
      if ((count % 100000) == 0) {
        elapsedSoFar := time.Now().Sub(start)
        fmt.Println("Checked", count, "passphrases within", elapsedSoFar)
      }
      response := <-channel
      if (response[2] == "Y") {
        fmt.Println("Passphrase:", response[0])
        fmt.Println("Address:", response[1])
        done = true
        break
      }
    }
    if (done) {
      break
    }
  }
}
