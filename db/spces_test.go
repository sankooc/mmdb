package db

import (
	"bytes"
	"fmt"
	"io/ioutil"
)

func insertMany(){
	data, _ := ioutil.ReadFile("../asset/req.insertMany.dump")
	fmt.Println(data)
	c := bytes.NewReader(data)
	w := &bytes.Buffer{}
	h := &Simple{collections: make(map[string]*MemoryCollection)}
	err := parse(c, w, h)
	if err != nil {
		fmt.Println(err)
	}
}

func insertOne(){
	data, _ := ioutil.ReadFile("../asset/req.insertOne.dump")
	fmt.Println(data)
	c := bytes.NewReader(data)
	w := &bytes.Buffer{}
	h := &Simple{collections: make(map[string]*MemoryCollection)}
	err := parse(c, w, h)
	if err != nil {
		fmt.Println(err)
	}
}
//func TestMaster(t *testing.T){
//	data, _ := ioutil.ReadFile("../asset/req.master.dump")
//	c := bytes.NewReader(data)
//	w := &bytes.Buffer{}
//	h := &Simple{collections: make(map[string]*MemoryCollection)}
//	err := parse(c, w, h)
//	if err != nil {
//		fmt.Println(err)
//	}
//}
//func TestRequest(t *testing.T) {
//	insertOne()
//}
