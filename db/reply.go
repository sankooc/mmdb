package db

import (
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

/*
BUILDINFO response for buildinfo: 1
*/
var BUILDINFO = bson.M{
	"version":          "4.4.0",
	"gitVersion":       "ad91a93a5a31e175f5cbf8c69561e788bbc55ce1",
	"allocator":        "system",
	"javascriptEngine": "mozjs",
	"sysInfo":          "deprecated",
	"buildEnvironment": bson.M{
		"distmod":  "",
		"distarch": "x86_64", "target_arch": "x86_64",
		"target_os": "macOS",
	},
	"bits":              64,
	"debug":             false,
	"maxBsonObjectSize": 2 ^ 24,
	"storageEngines": bson.A{
		"mmdb",
	},
	"ok": 1,
}

//var ISMASTER = markOk(bson.D{{"you", "127.0.0.1"}})

var ISMASTER = bson.M{
	"topologyVersion": bson.M{
		"processId": primitive.NewObjectID(),
		"counter":   uint64(0),
	},
	"maxBsonObjectSize":            16777216,
	"maxMessageSizeBytes":          48000000,
	"maxWriteBatchSize":            100000,
	"logicalSessionTimeoutMinutes": 30,
	"connectionId":                 59,
	"minWireVersion":               0,
	"maxWireVersion":               9,
	"readOnly":                     false,
	"ok":                           1,
	"ismaster":                     true,
	"localTime":                    primitive.NewDateTimeFromTime(time.Now()),
}

var WARNING = bson.M{
	"totalLinesWritten": 1,
	"log": bson.A{
		`*****MMDB****`,
	},
	"ok": 1,
}
