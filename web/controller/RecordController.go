package controller

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"dormon.net/qq/web/model"
	"dormon.net/qq/record_process"
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

	record_process.MsgChan <- msg

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

	record_process.PicChan <- pic

}
