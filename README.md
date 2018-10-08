## Go Vanity Script

<p align="center">
    <img src="https://github.com/ark-collective/go-vanity/blob/master/banner.png" />
</p>

### Install Go 1.10

1. Download Go 1.10 [here](https://golang.org/dl/) and install it.
2. Open up bash and go to the Go home directory: `cd $GOPATH/src`
3. Clone the repo `git clone git@github.com:ark-collective/go-vanity.git && cd go-vanity`
4. Start generating some fine addresses (see below)

### Install dep

Follow the instructions [here](https://github.com/golang/dep#installation)

### Install Dependencies

```
dep ensure
```

### Run Vanity Script

```
go run vanity.go -prefix ALEX -case-insensitive
go run vanity.go -p ALEX -i
```

### Options

## Prefix

`-p` or `-prefix`

Prefix to search for

**Prefix or Suffix required in order to do a search**

## Suffix

`-s` or `-suffix`

Suffix to search for

**Prefix or Suffix required in order to do a search**

## Prefix and Suffix

`-ps` or `-prefix-and-suffix`

Find address when both the Prefix and Suffix match [default=false]

## Entropy

`-e` or `-entropy`

Entropy for an address (E.g. 128 = 12 words, 256 = 24 words) [min=128 max=256 default=128]

## Address Format

`-a` or `-address-format`

First letter of address (E.g. 23 = A, 30 = D). This is to search for addresses based on network [default=23]

## Threads

`-t` or `-threads`

How many threads to use when searching for addresses. "0" enables a benchmark to try and work out the best value [default=100]

## WIF

`-w` or `-wif`

The network version [default=170]

## Case Insensitive

`-i` or `-case-insensitive`

Whether to check exact case or not [default=false]

## Count

`-c` or `-count`

Quantity of addresses to search for (milestone option doesn't work with more than 1 count) [default=1]

## File Output

`-o` or `-output`

Output results to file [default=results.txt]
