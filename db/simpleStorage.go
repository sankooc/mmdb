package db

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"strconv"
	"sync"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type MemoryCollection struct {
	docs map[string]bson.M
	mu   sync.RWMutex
}
type Simple struct {
	collections map[string]*MemoryCollection
	lastErr     interface{}
	mu          sync.RWMutex
}

func (s *Simple) listDatabases() bson.M {
	fmt.Println("simple storage")
	var dbinfos []bson.D

	dbinfos = append(dbinfos, bson.D{
		{"name", "test"},
		{"empty", true},
	})

	dbinfos = append(dbinfos, bson.D{
		{"name", "tmp"},
		{"empty", false},
		{"sizeOnDisk", 40960},
	})
	return bson.M{"ok": 1, "databases": dbinfos}
}

func (s *Simple) listCollections(section bson.M, dbname string) bson.M {
	return bson.M{
		"ok": 1,
		"cursor": bson.M{
			"id": uint64(0),
			"ns": "local.$cmd.listCollections",
			"firstBatch": bson.A{bson.M{
				"name": "usr",
				"type": "collection",
				"info": bson.M{
					"readOnly": false,
				},
				"idIndex": bson.M{
					"v": 2,
					"key": bson.M{
						"_id": 1,
					},
					"name": "_id_",
				},
			}},
		},
	}
}

func (s *Simple) command(utype string, cmd string, arg string) bson.M {
	Debug("cmd %s %s \r\n", utype, cmd)
	switch cmd {
	case "getLog":
		switch arg {
		case "*":
			return bson.M{"ok": 1, "names": []string{"startupWarnings"}}
		case "startupWarnings":
			return bson.M{"ok": 1, "totalLinesWritten": 0, "log": []string{}}
		default:
			return errSReply()
		}
	case "listDatabases":
		return s.listDatabases()
	case "replSetGetStatus":
		return okReply()
	case "shutdown":
		panic("shutdown requested")
	case "whatsmyuri":
		return okReply()
	case "ismaster", "isMaster":
		return ISMASTER
	case "getnonce":
		nonce := make([]byte, 32)
		_, err := rand.Reader.Read(nonce[:])
		if err != nil {
			return errSReply()
		}
		return bson.M{"ok": 1, "nonce": hex.EncodeToString(nonce)}
	case "ping":
		return okReply()
	case "buildinfo", "buildInfo":
		return BUILDINFO
	case "getFreeMonitoringStatus":
		return okReply()
	case "getCmdLineOpts":
		return okReply()
	case "listCollections":
		return bson.M{"ok": 1, "cursor": bson.M{
			"id":         uint64(0),
			"ns":         "local.$cmd.listCollections",
			"firstBatch": bson.A{},
		}}
	case "getLastError", "getlasterror":
		return okReply() // TODO
	default:
		return okReply()
	}

	return nil
}

func (s *Simple) getDb(db string, col string) *MemoryCollection {
	var cols *MemoryCollection = s.collections[col]
	if cols == nil {
		cols = &MemoryCollection{docs: make(map[string]bson.M)}
		s.collections[col] = cols
	}
	return cols
}

func (s *Simple) insert(db string, col string, doc bson.A) bson.M {
	size := len(doc)
	if size > 0 {
		collection := s.getDb(db, col)
		docs := collection.docs

		for _, dc := range doc {
			d := dc.(bson.M)
			//delete(d, "name")
			id := d["_id"]
			if id == nil {
				id = primitive.NewObjectID()
				d["_id"] = id
			}
			_id := id.(primitive.ObjectID).Hex()
			Debug(" %s \r\n", _id)
			docs[_id] = d
		}
		return bson.M{"n": size, "ok": 1}
	}
	return bson.M{"n": size, "ok": 0}
}

func (s *Simple) delete(db string, col string, opt bson.M) bson.M {
	collection := s.getDb(db, col)
	q := opt["q"].(bson.M)
	limit, _ := strconv.ParseInt(fmt.Sprint(opt["limit"]), 10, 32)
	n := 0
	docs := collection.docs
	for k, d := range docs {
		if n >= int(limit) && limit > 0 {
			break
		}
		m := isPatternMatch(d, q)
		if m == true {
			n += 1
			delete(docs, k)
		}
	}
	return bson.M{"ok": 1, "n": n}
}

func (s *Simple) query(db string, col string, query bson.M) bson.M {
	collection := s.getDb(db, col)
	docs := collection.docs
	matchs := make(bson.A, 0)
	for _, d := range docs {
		m := isPatternMatch(d, query)
		if m == true {
			matchs = append(matchs, d)
		}
	}
	return bson.M{
		"cursor": bson.M{
			"firstBatch": matchs,
			"id":         int64(0),
			"ns":         fmt.Sprintf("%s.%s", db, col),
		},
		"ok": 1,
	}
}

func asBsonM(v interface{}) (bson.M, error) {
	if v == nil {
		return nil, nil
	}
	switch b := v.(type) {
	case bson.D:
		return b.Map(), nil
	case bson.M:
		return b, nil
	}
	return nil, fmt.Errorf("cannot resolve %q to bson.M", v)
}

func (s *Simple) update(db string, col string, opt bson.M) bson.M {
	collection := s.getDb(db, col)
	docs := collection.docs
	rs := bson.M{"n": 0, "nModified": 0, "ok": 1}
	query := opt["q"].(bson.M)
	up := opt["u"].(bson.M)
	multi := opt["multi"] == true
	upsert := opt["upsert"] == true
	nModified := 0
	n := 0
	for _, d := range docs {
		m := isPatternMatch(d, query)
		if m == true && (multi || nModified <= 0) {
			err := updateDoc(d, up)
			if err != nil {
				nModified += 1
				n += 1
			}
		}
	}
	if nModified == 0 && upsert {
		n += 1
		// create
	}
	rs["nModified"] = nModified
	rs["n"] = n
	return rs
}
