package main

import (
	"fmt"
	"github.com/sankooc/mmdb/db"
	"os"
)

func main() {
	arg := os.Args
	port := "22222"
	if len(arg) >= 3 {
		if arg[1] == "--port" {
			port = arg[2]
		}
	}
	server, err := db.NewServerAddr(fmt.Sprintf(":%s", port), nil)
	if err != nil {
		panic(err)
	}
	server.Start()
	server.Wait()
}
