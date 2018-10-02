## Go Vanity Script

### Install dep

Follow the instructions [here](https://github.com/golang/dep#installation)

### Install Dependencies

```
dep ensure
```

### Run Vanity Script

```
go run vanity.go -prefix ALEX -entropy 128 -public-key-hash 23 -threads 100 -wif 170
```
