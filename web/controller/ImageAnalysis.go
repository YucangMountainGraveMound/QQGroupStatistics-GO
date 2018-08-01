package controller

import (
	"github.com/gin-gonic/gin"
	"github.com/olivere/elastic"
	"context"
	"net/http"
)

/**
{
	"size": 0,
	"query": {
		"bool": {
			"must_not": {
				"script": {
					"script": {
						"source": "doc['images'].values.length == 0"
					}
				}
			}
		}
	},
	"aggs": {
		"result":{
			"terms": {
				"size": 20,
				"field": "images"
			},
			"aggs": {
				"result": {
						"terms": {
							"size": 10,
							"field": "number"
					}
				}
			}
		}
	}
}
 */
func ImagesCountWithUser(c *gin.Context) {
	queryService := prepare(c)

	mustNotQuery := elastic.NewScriptQuery(elastic.NewScript("doc['images'].values.length == 0"))
	boolQuery := elastic.NewBoolQuery().MustNot(mustNotQuery)
	subAggs := elastic.NewTermsAggregation().Field("number").Size(10)
	aggs := elastic.NewTermsAggregation().Field("images").Size(20)
	aggs.SubAggregation("result", subAggs)

	searchResult, err := queryService.Query(boolQuery).Aggregation("result", aggs).Do(context.Background())
	if !handleError(err, c) {
		return
	}

	c.JSON(http.StatusOK, searchResult.Aggregations["result"])
}
