watch:
	fswatch -0 db | xargs -0 -I {} make test

test:
	cd db && go test
