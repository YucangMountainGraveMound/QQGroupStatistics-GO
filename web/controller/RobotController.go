package controller

import (
	"github.com/gin-gonic/gin"
	"strings"
	"github.com/jinzhu/now"
	"encoding/json"
	"strconv"
	"net/http"
	"github.com/olivere/elastic"
	"context"
	"dormon.net/qq/web/model"
	"github.com/jinzhu/gorm"
)

type Result struct {
	Item []Buckets `json:"buckets"`
}

type Buckets struct {
	Key      string `json:"key"`
	DocCount int    `json:"doc_count"`
}

func RobotMessageCount(c *gin.Context) {
	param := c.Param("param")


}

/**
{
	"size": 0,
	"query": {
        "range" : {
            "date" : {
                "gte" : "xxxx",
                "lt" :  "xxxx"
            }
        }
    },
	"aggs": {
		"result": {
			"terms": {
				"field": "number"
			}
		}
	}
}
 */
func mostTalker(c *gin.Context) {
	queryService := prepare(c)

	query := elastic.NewRangeQuery("date").Gte(now.BeginningOfDay())
	aggs := elastic.NewTermsAggregation().Field("number").Size(10)

	searchResult, err := queryService.Query(query).Aggregation("result", aggs).Do(context.Background())
	if !handleError(err, c) {
		return
	}

	var result Result
	err = json.Unmarshal(*searchResult.Aggregations["result"], &result)
	if !handleError(err, c) {
		return
	}

	if len(result.Item) != 0 {
		resultString := "今天" + result.Item[0].Key + "最水, 水了" + strconv.Itoa(result.Item[0].DocCount) + "条"
		c.String(http.StatusOK, resultString)
	} else {
		c.String(http.StatusOK, "还没有人在水，赶紧来个人水啊！")
	}
}

/**
{
	"size": 0,
	"query": {
        "range" : {
            "date" : {
                "gte" : "xxxx",
                "lt" :  "xxxx"
            }
        }
    },
	"aggs": {
		"result": {
			"terms": {
				"field": "number"
			}
		}
	}
}
 */
func ranking(c *gin.Context) {
	queryService := prepare(c)

	query := elastic.NewRangeQuery("date").Gte(now.BeginningOfDay())
	aggs := elastic.NewTermsAggregation().Field("number").Size(10)

	searchResult, err := queryService.Query(query).Aggregation("result", aggs).Do(context.Background())
	if !handleError(err, c) {
		return
	}

	var result Result
	err = json.Unmarshal(*searchResult.Aggregations["result"], &result)
	if !handleError(err, c) {
		return
	}

	if len(result.Item) != 0 {
		i := 0
		resultString := ""
		for _, v := range result.Item {
			i ++
			resultString += "第" + strconv.Itoa(i) + "名:" + v.Key + ",水了" + strconv.Itoa(v.DocCount) + "条\n"
		}
		c.String(http.StatusOK, resultString)
	} else {
		c.String(http.StatusOK, "还没有人在水，赶紧来个人水啊！")
	}
}

func ShootList(c *gin.Context, today bool) {

	queryService := prepare(c)

	include, err := model.GetDictionaryByKey("shot_list")
	if !handleError(err, c) {
		return
	}
	includeString := include.Value

	exclude, err := model.GetDictionaryByKey("shot_list" + "_exclude")
	var excludeString string
	if err == gorm.ErrRecordNotFound {
		excludeString = ""
	} else {
		if !handleError(err, c) {
			return
		}
	}
	excludeString = exclude.Value

	_includes := strings.Split(includeString, ",")
	includes := make([]interface{}, len(_includes))
	for i, v := range _includes {
		includes[i] = v
	}
	_excludes := strings.Split(excludeString, ",")
	excludes := make([]interface{}, len(_excludes))
	for i, v := range _excludes {
		excludes[i] = v
	}

	mustTermsQuery := elastic.NewTermsQuery("message", includes...)
	mustNotTermsQuery := elastic.NewTermsQuery("message", excludes...)
	query := elastic.NewBoolQuery().Must(mustTermsQuery).MustNot(mustNotTermsQuery)
	if today {
		query = query.Must(elastic.NewRangeQuery("date").Gte(now.BeginningOfDay()))
	}
	aggs := elastic.NewTermsAggregation().Field("number")
	searchResult, err := queryService.Query(query).Aggregation("result", aggs).Do(context.Background())
	handleError(err, c)

	var result Result
	err = json.Unmarshal(*searchResult.Aggregations["result"], &result)
	if !handleError(err, c) {
		return
	}

	if len(result.Item) != 0 {
		i := 0
		resultString := ""
		for _, v := range result.Item {
			i ++
			resultString += "第" + strconv.Itoa(i) + "名:" + v.Key + ",应该枪毙" + strconv.Itoa(v.DocCount) + "回\n"
		}
		c.String(http.StatusOK, resultString)
	} else {
		c.String(http.StatusOK, "哇，今天居然没有需要枪毙的")
	}
}

func unknownContent(c *gin.Context) {
	c.String(http.StatusOK, "说杰宝啥呢，当老子是人啊！")
}
