package controller

import (
	"context"
	"net/http"

	"dormon.net/qq/config"

	"github.com/gin-gonic/gin"
	"github.com/olivere/elastic"
)

// Histogram 历史视图
/**
{
	"size": 0,
	"aggs": {
	    "result": {
	    	"date_histogram": {
	    		"field": "date",
	    		"interval": "1d",
	    		"min_doc_count": 0
	    	}
	    }
	}
}
*/
func Histogram(c *gin.Context) {
	queryService := prepare(c)

	interval := c.DefaultQuery("interval", "1d")

	aggs := elastic.NewDateHistogramAggregation().Field("date").Interval(interval).MinDocCount(0)
	searchResult, err := queryService.Aggregation("result", aggs).Do(context.Background())
	handleError(err, c)

	c.JSON(http.StatusOK, searchResult.Aggregations["result"])
}

// MessageCountByUser 每个用户的消息数量以及图片数量
/**
{
	"size": 0,
	"aggs": {
		"account": {
			"terms":{
				"field": "number"
			},
			"aggs": {
				"image_count": {
					"value_count": {
						"field": "images"
					}
				}
			}
		}
	}
}
*/
func MessageCountByUser(c *gin.Context) {
	queryService := prepare(c)

	subAggs := elastic.NewValueCountAggregation().Field("images")
	aggs := elastic.NewTermsAggregation().Field("number").SubAggregation("image_count", subAggs).Size(len(config.Config().Account))
	searchResult, err := queryService.Aggregation("result", aggs).Do(context.Background())
	handleError(err, c)

	c.JSON(http.StatusOK, searchResult.Aggregations["result"])
}

// MessageCountBySpecificTime 根据特定时间统计
/**
{
	"size": 0,
	"aggs": {
	    "dayOfWeek": {
		        "terms": {
		            "script": "doc['date'].date.dayOfWeek",
		            "order": {
		                "_term": "asc"
		            }
		        },
		        "aggs": {
		            "hourOfDay": {
				        "terms": {
				            "script": "doc['date'].date.hourOfDay",
				            "order": {
				                "_term": "asc"
				            },
				            "size": 100
				        }
				    }
		        }
	    }
	}
}
*/
func MessageCountBySpecificTime(c *gin.Context) {
	queryService := prepare(c)

	by := c.DefaultQuery("by", "hourOfDay")
	of := c.DefaultQuery("of", "dayOfWeek")

	subAggs := elastic.NewTermsAggregation().Script(elastic.NewScript("doc['date'].date."+by)).Order("_term", true).Size(100)
	aggs := elastic.NewTermsAggregation().Script(elastic.NewScript("doc['date'].date."+of)).Order("_term", true)
	aggs.SubAggregation("result", subAggs)

	searchResult, err := queryService.Aggregation("result", aggs).Do(context.Background())
	handleError(err, c)

	c.JSON(http.StatusOK, searchResult.Aggregations["result"])
}

/**
{
	"size": 0,
	"aggs": {
		"dayOfWeek": {
			"terms": {
				"script": "doc['date'].date.dayOfWeek==3&&doc['date'].date.hourOfDay==2"
			},
			"aggs": {

				"result": {
					"terms": {
						"field": "message"
					}
				}
			}
		}
	}
}
*/
func MessagesBySepcificTime(c *gin.Context) {

}

/**
{
	"size": 0,
    "aggs" : {
        "result" : {
            "terms" : {
                "field" : "number",
                "size": 100
            },
            "aggs" : {
		        "result" : {
		            "cardinality" : {
		                "script": {
		                    "lang": "painless",
		                    "source": "doc['date'].date.dayOfYear"
		                }
		            }
		        }
		    }
        }
    }
}
*/
func UsersByDay(c *gin.Context) {
	queryService := prepare(c)

	subAggs := elastic.NewCardinalityAggregation().Script(elastic.NewScript("doc['date'].date.dayOfYear"))
	aggs := elastic.NewTermsAggregation().Field("number").Size(100)
	aggs.SubAggregation("result", subAggs)

	searchResult, err := queryService.Aggregation("result", aggs).Do(context.Background())
	handleError(err, c)

	c.JSON(http.StatusOK, searchResult.Aggregations["result"])
}

/**
{
	"size": 0,
	"query": {
		"bool": {
			"must": [{
				"range": {
					"interval": {
						"lte": 60,
						"gte": 0
					}
				}
			}]
		}
	},
		"aggs": {
		"result": {
			"terms": {
				"field": "number",
				"size": 100
			},
			"aggs": {
				"result": {
					"histogram": {
						"field": "interval",
						"interval": 1
					},
					"aggs": {
						"result": {
							"value_count": {
								"field": "message"
							}
						}
					}
				}
			}
		}
	}
}
*/
func MessageHabit(c *gin.Context) {
	queryService := prepare(c)

	intervalRangeQuery := elastic.NewRangeQuery("interval").Lte(60).Gte(0)
	boolQuery := elastic.NewBoolQuery().Must(intervalRangeQuery)
	aggs2 := elastic.NewValueCountAggregation().Field("message")
	aggs1 := elastic.NewHistogramAggregation().Field("interval").Interval(1)
	aggs := elastic.NewTermsAggregation().Field("number").Size(100)
	aggs1.SubAggregation("result", aggs2)
	aggs.SubAggregation("result", aggs1)

	searchResult, err := queryService.Query(boolQuery).Aggregation("result", aggs).Do(context.Background())
	if !handleError(err, c) {
		return
	}

	c.JSON(http.StatusOK, searchResult.Aggregations["result"])
}
