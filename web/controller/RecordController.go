package controller

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"github.com/sirupsen/logrus"
)

type Message struct {
	ClassName string `json:"className"`
	FriendUin string `json:"friendUin"`
	Message   string `json:"message"`
	SelfUin   string `json:"selfUin"`
	SenderUin string `json:"senderUin"`
	Time      string `json:"time"`
	UniSeq    string `json:"uniSeq"`
}

type Picture struct {
	PicUrl string `json:"picUrl"`
	UniSeq string `json:"uniSeq"`
}

func RecordMessage(c *gin.Context) {
	var err error
	var msg Message
	err = c.BindJSON(&msg)
	if !handleError(err, c) {
		c.JSON(http.StatusBadRequest, gin.H{
			"dormon":  "",
			"message": "Json content may not right",
		})
		return
	}

	logrus.Info(msg)

}

func RecordPicture(c *gin.Context) {
	var err error
	var pic Picture
	err = c.BindJSON(&pic)
	if !handleError(err, c) {
		c.JSON(http.StatusBadRequest, gin.H{
			"dormon":  "",
			"message": "Json content may not right",
		})
		return
	}

	logrus.Info(pic)

}
