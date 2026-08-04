package db

import "go.mongodb.org/mongo-driver/v2/bson"

func cardtRankFusionDocument(query string, queryVector []float32, limit int, excludeID string) bson.D {
	vectorSearch := bson.D{
		{Key: "index", Value: "text_embedding"},
		{Key: "path", Value: "textEmbedding"},
		{Key: "exact", Value: false},
		{Key: "numCandidates", Value: limit * 2 * 10},
		{Key: "queryVector", Value: queryVector},
		{Key: "limit", Value: limit * 2},
	}
	if excludeID != "" {
		vectorSearch = append(vectorSearch, bson.E{Key: "filter", Value: bson.D{
			{Key: "id", Value: bson.D{
				{Key: "$ne", Value: excludeID},
			}},
		}})
	}
	vectorPipeline := bson.A{
		bson.D{{Key: "$vectorSearch", Value: vectorSearch}},
	}

	textPipeline := bson.A{
		bson.D{
			{Key: "$search", Value: bson.D{
				{Key: "index", Value: "text_search"},
				{Key: "text", Value: bson.D{
					{Key: "query", Value: query},
					{Key: "path", Value: "text"},
					// {Key: "fuzzy", Value: bson.D{}},
				}},
			}},
		},
	}
	if excludeID != "" {
		textPipeline = append(textPipeline, bson.D{
			{Key: "$match", Value: bson.D{
				{Key: "id", Value: bson.D{
					{Key: "$ne", Value: excludeID},
				}},
			}},
		})
	}
	textPipeline = append(textPipeline, bson.D{{Key: "$limit", Value: limit * 2}})

	return bson.D{
		{Key: "$rankFusion", Value: bson.D{
			{Key: "input", Value: bson.D{
				{Key: "pipelines", Value: bson.D{
					{Key: "vectorPipeline", Value: vectorPipeline},
					{Key: "textPipeline", Value: textPipeline},
				}},
			}},
			{Key: "combination", Value: bson.D{
				{Key: "weights", Value: bson.D{
					{Key: "vectorPipeline", Value: vectorPipelineWeight},
					{Key: "textPipeline", Value: textPipelineWeight},
				}},
			}},
		}},
	}
}
