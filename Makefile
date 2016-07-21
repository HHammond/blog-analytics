
PRODUCT="analytics"

.pre-build:
	go get -u gopkg.in/reform.v1
	go get -u github.com/mattn/go-sqlite3
	touch .pre-build

.PHONY: build
build: .pre-build
	go generate 
	go build

.PHONY: build-arm
build-arm: .pre-build
	GOOS=linux GOARCH=arm GOARM=7 CGO_ENABLED=0 go build

.PHONY: clean
clean:
	- rm -f $(PRODUCT)
	- find . -name "*_reform.go" -delete
	- rm -f .pre-build
