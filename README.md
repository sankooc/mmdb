[![Build Status](https://api.travis-ci.org/sankooc/mmdb.svg)](http://travis-ci.org/sankooc/mmdb)


## MMDB

In-*memory* *MongoDB* Server written with pure golang  to use in place of mongodb in your unit tests



## Usage



### Code

```golang


import (
	"fmt"
	"github.com/sankooc/mmdb/db"
	"context"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)
func mockMongo() error {
	server, err := db.NewServerAddr(":22222", nil)
	if err != nil {
		return err
	}
	server.Start()
	fmt.Println("mongodb_started")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	client, err := mongo.Connect(ctx, options.Client().ApplyURI("mongodb://127.0.0.1:22222"))
	defer cancel()
	if err != nil {
		return err
	}
	collection := client.Database("test").Collection("usr")
	doc := bson.M{"name": "sankooc", "age": 30}
	_, err = collection.InsertOne(ctx, doc)
	if err != nil {
		return err
	}
	filter := bson.M{}
	err = collection.FindOne(ctx, filter).Decode(&doc)
	if err != nil {
		return err
	}
	fmt.Println(doc)
	return nil
}
```



### Cli



```golang
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

```

connect witch mongo-cli

` mongo --port 22222`



## Status



#### Collection Command

| opt  | -    |
| ---- | ---- |
| insertOne | ✔️ |
| find | ✔️ |
| update | ✔️ |
| insert | ✔️ |
| deleteOne | ✔️ |
| insertMany | ✔️ |
| count | ❌ |
| countDocuments | ❌ |
| findOne | ❌ |
| deleteMany | ❌ |
| replaceOne | ❌ |
| findAndModify | ❌ |
| findOneAndReplace | ❌ |
| findOneAndUpdate | ❌ |
| save | ❌ |
| drop | ❌ |
| updateMany | ❌ |
| remove | ❌ |
| bulkWrite | ❌ |
| dropIndex | ❌ |
| getDB | ❌ |
| updateOne | ❌ |
| mapReduce | ❌ |
| getIndexKeys | ❌ |
| getWriteConcern | ❌ |
| stats | ❌ |
| convertToSingleObject | ❌ |
| estimatedDocumentCount | ❌ |
| getIndexSpecs | ❌ |
| storageSize | ❌ |
| exists | ❌ |
| getIndexes | ❌ |
| explain | ❌ |
| getIndices | ❌ |
| createIndex | ❌ |
| hideIndex | ❌ |
| renameCollection | ❌ |
| createIndexes | ❌ |
| getName | ❌ |
| initializeOrderedBulkOp | ❌ |
| totalIndexSize | ❌ |
| dataSize | ❌ |
| getPlanCache | ❌ |
| totalSize | ❌ |
| findOneAndDelete | ❌ |
| getQueryOptions | ❌ |
| aggregate | ❌ |

#### Update Operators


| opt  | -    |
| ---- | ---- |
| $set| ✔️ |
| $unset| ✔️ |
| $setOnInsert| ❌ |
| $inc| ❌ |
| $min| ❌ |
| $max| ❌ |
| $mul| ❌ |
| $rename| ❌ |
| $currentDate| ❌ |

#### Query Operators

|  opt |   -  |
| ---- | ---- |
| $eq | ❌ |
| $gt | ❌ |
| $lt | ❌ |
| $gte | ❌ |
| $lte | ❌ |
| $exists | ❌ |
