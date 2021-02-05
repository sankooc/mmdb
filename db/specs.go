package db

import "go.mongodb.org/mongo-driver/bson"

// MongoMessage mongo common message
type MongoMessage struct {
	Length     uint32
	RequestID  uint32
	ResponseTo uint32
	OpCode     uint32
	Data       []byte
}

type M2013 struct {
	Message  *MongoMessage
	FlagBits int
	Meta     []Section
	Sections []Section
}
type M2004 struct {
	Message              *MongoMessage
	Flag                 int
	FullCollectionName   string
	NumberToSkip         uint32
	NumberToReturn       uint32
	Doc                  bson.D
	ReturnFieldsSelector bson.D
}

func (m *M2004) Command() (cmd string, arg interface{}) {
	for _, kv := range m.Doc {
		return kv.Key, kv.Value
	}
	return "", nil
}

func (m *M2004) Get(key string) (interface{}, bool) {
	for _, kv := range m.Doc {
		if kv.Key == key {
			return kv.Value, true
		}
	}
	return nil, false
}

type M2013Reply struct {
	Message  *MongoMessage
	FlagBits int
	sections []Section
}
type M2004Reply struct {
	Message        *MongoMessage
	Flag           uint32
	CursorID       uint64
	StartingFrom   uint32
	NumberReturned uint32
	Docs           []interface{}
}

type Body1 bson.M

type Body2 struct {
	doc    bson.M
	indent string
}
type Section interface{}

//type Header struct {
//	Length     int32
//	RequestID  int32
//	ResponseTo int32
//	OpCode     OpCode
//	Contents   []byte
//}

type MMOpt struct {
	hasCheckSum bool
	engine      MongoEngine
}

/*
 spec: https://github.com/mongodb/specifications/blob/master/source/crud/crud.rst

*/
type MongoEngine interface {
	listDatabases() bson.M
	listCollections(section bson.M, dbname string) bson.M
	command(utype string, cmd string, arg string) bson.M
	// https://github.com/mongodb/specifications/blob/master/source/enumerate-collections.rst
	insert(db string, col string, doc bson.A) bson.M
	query(db string, col string, query bson.M, singleBatch bool) bson.M
	delete(db string, col string, opt bson.M) bson.M
	update(db string, col string, opt bson.M) bson.M
	count(db string, col string, opt bson.M) bson.M
}
