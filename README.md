## Go Vanity Script

### Install dep

Follow the instructions [here](https://github.com/golang/dep#installation)

### Install Dependencies

```
dep ensure
```

### Run Vanity Script

```
go run vanity.go -prefix ALEX -entropy 128 -address-format 23 -threads 100 -wif 170
go run vanity.go -p ALEX -e 128 -a 23 -t 100 -w 170
```
