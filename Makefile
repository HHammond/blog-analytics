
PRODUCT="analytics"

.pre-build:
	go get -u github.com/mattn/go-sqlite3
	go get -u github.com/caarlos0/env
	touch .pre-build

.PHONY: build
build: .pre-build
	go build

.PHONY: install
install: .pre-build
	go install

.PHONY: clean
clean:
	- rm -f $(PRODUCT)
	- rm -f .pre-build
	- find . -name "*.db" -delete
	- find . -name "*.db-journal" -delete
