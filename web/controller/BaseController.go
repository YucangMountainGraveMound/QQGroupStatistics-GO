package controller

import (
	"net/http"
	"reflect"
	"strconv"

	"dormon.net/qq/es"

	"github.com/appleboy/gin-jwt"
	"github.com/gin-gonic/gin"
	"github.com/jinzhu/now"
	"github.com/olivere/elastic"
	"github.com/sirupsen/logrus"
)

// CommonParams 提取出共有的请求参数
type CommonParams struct {
	CurrentAccount string
	TargetAccount  string
	TimeGTE        interface{}
	TimeLTE        interface{}
	Size           int
	From           int
	To             int
}

var (
	cp       *CommonParams
	esClient *elastic.Client
)

// extractParams 提取公用参数
func (cp *CommonParams) extractParams(c *gin.Context) {
	var err error
	a := jwt.ExtractClaims(c)["id"]
	if reflect.TypeOf(a) == reflect.TypeOf("") {
		cp.CurrentAccount = jwt.ExtractClaims(c)["id"].(string)
	}
	cp.TargetAccount = c.DefaultQuery("account", "_all")
	cp.TimeGTE = c.DefaultQuery("gte", "_all")
	if cp.TimeGTE != "_all" {
		cp.TimeGTE, err = now.Parse(cp.TimeGTE.(string))
		handleError(err, c)
	}
	cp.TimeLTE = c.DefaultQuery("lte", "_all")
	if cp.TimeLTE != "_all" {
		cp.TimeLTE, err = now.Parse(cp.TimeLTE.(string))
		handleError(err, c)
	}
	cp.Size, _ = strconv.Atoi(c.DefaultQuery("size", "0"))
	handleError(err, c)
	cp.From, _ = strconv.Atoi(c.DefaultQuery("from", "0"))
	handleError(err, c)
}

// 统一设置es相关前置操作
func prepare(c *gin.Context) (queryService *elastic.SearchService) {
	esClient = es.ElasticClient()
	cp = &CommonParams{}
	cp.extractParams(c)

	queryService = esClient.Search()
	query := elastic.NewBoolQuery()

	// 检索特定账号
	if cp.TargetAccount != "_all" {
		if cp.CurrentAccount == cp.TargetAccount {
			query = query.Must(elastic.NewTermQuery("number", cp.TargetAccount))
		} else {
			// 暂时不允许查看他人的数据
			// TODO:可以做个授权
			c.JSON(http.StatusBadRequest, gin.H{
				"message": "Out of privacy concern, other's info not accessible.",
			})
		}
	}

	// 检索特定时间
	rangeQuery := elastic.NewRangeQuery("date")
	if cp.TimeGTE != "_all" {
		rangeQuery = rangeQuery.Gte(cp.TimeGTE)
	}
	if cp.TimeLTE != "_all" {
		rangeQuery = rangeQuery.Lte(cp.TimeLTE)
	}
	query = query.Must(rangeQuery)

	queryService = queryService.Query(query).Size(cp.Size).From(cp.From)

	return
}

// 错误处理
func handleError(err error, c *gin.Context) bool {
	if err != nil {
		logrus.Errorf("Web services error: %s", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "error",
			"dormon":  "😱😱😱😱😱😱",
			"message": err.Error(),
		})
		return false
	}
	return true
}

// 是否是管理员
// TODO: 写死先，下辈子有空再改
func isAdmin(c *gin.Context) bool {
	a := jwt.ExtractClaims(c)["id"]
	if reflect.TypeOf(a) == reflect.TypeOf("") {
		a = jwt.ExtractClaims(c)["id"].(string)
	}
	if a == "422680319" {
		return true
	} else {
		c.JSON(http.StatusForbidden, gin.H{
			"message": "You are not allowed except you are an administrator.",
		})
		return false
	}
}
