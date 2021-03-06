package db

import (
	"testing"
	"github.com/stretchr/testify/assert"
	"go.mongodb.org/mongo-driver/bson"
)

func Test(t *testing.T) {
	storage := &Simple{collections: make(map[string]*MemoryCollection)}
	var db = "test"
	var col = "usr"
	{
		rs := storage.query(db, col, bson.M{}, false)
		assert.Equal(t, rs["ok"], 1)
		list := rs["cursor"].(bson.M)["firstBatch"].(bson.A)
		assert.Equal(t, len(list), 0)
	}
	{
		rs := storage.insert(db, col, bson.A{bson.M{"name": "aa"}, bson.M{"name": "aa"}, bson.M{"name": "cc"}, bson.M{"name": "aa"}})
		assert.Equal(t, rs["n"], 4)
		assert.Equal(t, rs["ok"], 1)
	}
	{
		rs := storage.query(db, col, bson.M{}, false)
		assert.Equal(t, rs["ok"], 1)
		list := rs["cursor"].(bson.M)["firstBatch"].(bson.A)
		assert.Equal(t, len(list), 4)
	}
	{
		rs := storage.count(db, col, bson.M{ "query": bson.M{}})
		assert.Equal(t, rs["ok"], 1)
		assert.Equal(t, rs["n"].(int), 4)
	}
	{
		rs := storage.query(db, col, bson.M{}, true)
		assert.Equal(t, rs["ok"], 1)
		list := rs["cursor"].(bson.M)["firstBatch"].(bson.A)
		assert.Equal(t, len(list), 1)
	}
	{
		rs := storage.delete(db, col, bson.M{"q": bson.M{"name": "aa"}, "limit": 2})
		assert.Equal(t, rs["ok"], 1)
		assert.Equal(t, rs["n"], 2)
	}
	{
		rs := storage.query(db, col, bson.M{}, false)
		assert.Equal(t, rs["ok"], 1)
		list := rs["cursor"].(bson.M)["firstBatch"].(bson.A)
		assert.Equal(t, len(list), 2)
	}
	{
		rs := storage.query(db, col, bson.M{"name": "cc"}, false)
		item := rs["cursor"].(bson.M)["firstBatch"].(bson.A)[0].(bson.M)
		assert.Equal(t, item["name"], "cc")
	}
}
