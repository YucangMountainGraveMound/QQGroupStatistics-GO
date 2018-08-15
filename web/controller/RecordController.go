package controller

import (
	"dormon.net/qq/web/model"
	"github.com/gin-gonic/gin"
	"net/http"
	"dormon.net/qq/utils"
	"io/ioutil"
	"github.com/sirupsen/logrus"
	"encoding/json"
	"dormon.net/qq/config"
	"strconv"
)

func CoolQ(c *gin.Context) {
	hmac := c.Request.Header["X-Signature"][0][5:]
	body, _ := ioutil.ReadAll(c.Request.Body)
	if utils.HMAC(body, config.Config().CoolQSecret) != hmac {
		c.String(http.StatusBadRequest, "Invalid token.")
		return
	}

	var record model.CoolQRecord

	logrus.Infoln(string(body))
	err := json.Unmarshal(body, &record)

	if err != nil {
		logrus.Error(err)
	}

	if strconv.Itoa(record.GroupId) == config.Config().SpecificGroup {
		model.CreateRecordFromCoolQMessage(record)
	}
}
