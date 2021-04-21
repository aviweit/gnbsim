all: build

build: update
	go build

update:
	go get -u

clean:
	rm -f gnb-gw go.sum
	git checkout go.mod
