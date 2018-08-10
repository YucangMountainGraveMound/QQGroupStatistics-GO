package controller

import (
	"context"
	"net/http"
	"strconv"
	"strings"

	"dormon.net/qq/errors"
	"dormon.net/qq/web/model"

	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"
	"github.com/olivere/elastic"
)

// WordFrequency 整体词频统计
/**
{
	"size": 0,
	"aggs": {
	    "result": {
	    	"terms": {
	    		"field": "message",
	    		"exclude": "[\u4E00-\u9FA5]",
	    		"size": 200
	    	}
	    }
	}
}
*/
// TODO: 分词数量是否有优雅点的实现方法，没搞明白英文分词数，还得研究一波
func WordFrequency(c *gin.Context) {
	queryService := prepare(c)

	// 去除单个汉字
	aggs := elastic.NewTermsAggregation().Field("message").Size(200)

	// 分词数
	wordNum, err := strconv.Atoi(c.DefaultQuery("word_num", "2"))
	handleError(err, c)

	wordRegexp := ""
	// 语言类型
	lang := c.DefaultQuery("lang", "zh")
	if lang == "zh" {
		for i := 0; i < wordNum; i++ {
			wordRegexp += "[\u4E00-\u9FA5]"
		}
	}

	aggs = aggs.Include(wordRegexp)
	searchResult, err := queryService.Aggregation("result", aggs).Do(context.Background())

	c.JSON(http.StatusOK, searchResult.Aggregations["result"])
}

/**
{
	"size": 0,
    "query": {
		"bool": {
			"must": {
				"terms": {
					"message": ["膜", "蛤","江","泽","民","暴力","蛙","青蛙","黑框","粉丝","+1","续"]
				}
			},
			"must_not": {
				"terms": {
					"message": ["马", "王","难民", "ryf", "分享"]
				}
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
// MessageByTerms 根据Terms对message进行统计
// @query key => Dictionaries表中相应的key字段
// @query show => 是否返回hit内容 便于调试
func MessageByTerms(c *gin.Context) {
	queryService := prepare(c)

	key := c.DefaultQuery("key", "")
	show := c.DefaultQuery("show", "false")

	if key == "" {
		handleError(errors.New("no key!"), c)
		return
	}

	include, err := model.GetDictionaryByKey(key)
	if !handleError(err, c) {
		return
	}
	includeString := include.Value

	exclude, err := model.GetDictionaryByKey(key + "_exclude")
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
	aggs := elastic.NewTermsAggregation().Field("number")
	searchResult, err := queryService.Query(query).Aggregation("result", aggs).Do(context.Background())
	handleError(err, c)

	if show == "true" {
		c.JSON(http.StatusOK, searchResult)
	} else {
		c.JSON(http.StatusOK, searchResult.Aggregations["result"])
	}
}
