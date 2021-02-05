tmp: clear
	mkdir -p target

clear:
	rm -rf target

build: tmp
	go build -o target/mmdb

coverage: tmp
	go test github.com/sankooc/mmdb/db -coverprofile ./cover.out && go tool cover -html=./cover.out

watch:
	fswatch -0 db | xargs -0 -I {} make test

test:
	cd db && go test
