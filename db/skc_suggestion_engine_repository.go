package db

import (
	"go.mongodb.org/mongo-driver/bson"
)

func Test() {
	doc := bson.D{{"title", "Record of a Shriveled Datum"}, {"text", "No bytes, no problem. Just insert a document, in MongoDB"}}
	skcSuggestionEngineDeckListCollection.InsertOne(ctx, doc)
}
