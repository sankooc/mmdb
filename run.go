package main

import (
	"github.com/sankooc/mmdb/db"
	"log"
	"os"
)

//
func die(err error) {
	if err != nil {
		log.Println(err)
		os.Exit(1)
	}
	os.Exit(0)
}

func main() {
	//tes := bson.A{bson.M{"qux":10}, bson.M{"qux":12345}}
	//b, err := bson.Marshal(tes)
	//if err != nil {
	//	panic(err)
	//}
	//fmt.Printf("size %d \r", len(b))
	server, err := db.NewServerAddr(":22222", nil)
	if err != nil {
		die(err)
	}
	server.Start()
	server.Wait()
	//if err != nil {
	//	die(err)
	//}
}
