package controller

import (
	"github.com/gin-gonic/gin"
	"github.com/olivere/elastic"
	"net/http"
	"context"
)

/**
{
	"_source": ["number", "at"],
	"size": 1000,
	"query": {
		"bool" :{
			"must_not": {
				"term": {
					"at": ""
				}
			}
		}
	}
}
 */
func UserAt(c *gin.Context) {
	queryService := prepare(c)

	fsc := elastic.NewFetchSourceContext(true).Include("number", "at")

	query := elastic.NewTermQuery("at", "")
	boolQuery := elastic.NewBoolQuery().MustNot(query)

	searchResult, err := queryService.Index("qq").Size(1000).Query(boolQuery).FetchSourceContext(fsc).Do(context.Background())
	handleError(err, c)

	c.JSON(http.StatusOK, searchResult.Hits.Hits)
}
