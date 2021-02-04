package main

import (
	"github.com/sankooc/mmdb/db"
)

func main() {
	server, err := db.NewServerAddr(":22222", nil)
	if err != nil {
		panic(err)
	}
	server.Start()
	server.Wait()
}
