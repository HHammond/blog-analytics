
PRODUCT="analytics"

.pre-build:
	go get -u github.com/mattn/go-sqlite3
	touch .pre-build

.PHONY: build
build: .pre-build
	go build 

.PHONY: build-arm
build-arm: .pre-build
	GOOS=linux GOARCH=arm GOARM=7 CGO_ENABLED=0 CC=gcc go build

.PHONY: clean
clean:
	- rm -f $(PRODUCT)
	- find . -name "*_reform.go" -delete
	- rm -f .pre-build
	- find . -name "*.db" -delete
	- find . -name "*.db-journal" -delete
