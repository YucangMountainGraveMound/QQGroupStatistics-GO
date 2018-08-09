package controller

import (
	"dormon.net/qq/web/model"
	"github.com/gin-gonic/gin"
	"net/http"
)

func RecordMessage(c *gin.Context) {
	var err error
	var msg model.Message
	err = c.BindJSON(&msg)
	if !handleError(err, c) {
		c.JSON(http.StatusBadRequest, gin.H{
			"dormon":  "",
			"message": "Json content may not right",
		})
		return
	}

	model.CreateRecordFromXposedMessage(msg)

}

func RecordPicture(c *gin.Context) {
	var err error
	var pic model.Picture
	err = c.BindJSON(&pic)
	if !handleError(err, c) {
		c.JSON(http.StatusBadRequest, gin.H{
			"dormon":  "",
			"message": "Json content may not right",
		})
		return
	}

	model.CreateRecordFromXposedPicture(pic)
}
