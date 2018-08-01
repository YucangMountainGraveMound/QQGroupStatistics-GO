package controller

import (
	"github.com/gin-gonic/gin"
	"dormon.net/qq/web/model"
	"net/http"
	"github.com/jinzhu/gorm"
	"strings"
	"fmt"
)

// SetDictionary 设置message terms查询时使用的字典
func SetDictionary(c *gin.Context) {
	var err error

	if !isAdmin(c) {
		return
	}

	key := c.PostForm("key")
	dict, _ := c.GetPostFormArray("dict")

	if key == "" || dict == nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"dormon":  "😵😵😵😵😵😵",
			"message": "params [key] or [dict] is required",
		})
		return
	}

	d, err := model.GetDictionaryByKey(key)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			d.Key = key
			d.Value = strings.Replace(strings.Trim(fmt.Sprint(dict), "[]"), " ", ",", -1)
			err = d.FirstOrCreate()
			if !handleError(err, c) {
				return
			}
		} else {
			if !handleError(err, c) {
				return
			}
		}
	} else {
		d.Value = strings.Replace(strings.Trim(fmt.Sprint(dict), "[]"), " ", ",", -1)
		err = d.Update()
		if !handleError(err, c) {
			return
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"dormon":  "🍻🍻🍻🍻🍻🍻",
		"message": "ok",
	})
}

// GetDictionary 获取message terms查询时使用的字典
func GetDictionary(c *gin.Context) {
	key := c.DefaultQuery("key", "")

	if key == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"dormon":  "😵😵😵😵😵😵",
			"message": "params [key] is required.",
		})
		return
	}

	s, err := model.GetDictionaryByKey(key)
	if !handleError(err, c) {
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"dormon":  "🍻🍻🍻🍻🍻🍻",
		"message": strings.Replace(strings.Trim(fmt.Sprint(s.Value), "[]"), " ", ",", -1),
	})
}
