package controller

import (
	"dormon.net/qq/web/model"

	"github.com/appleboy/gin-jwt"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"golang.org/x/crypto/bcrypt"
	"net/http"
	"strings"
)

// Authenticator Jwt验证用户信息
func Authenticator(username, password string, c *gin.Context) (interface{}, bool) {
	userInfo, err := model.FindUserByAccount(username)
	if err != nil {
		return "", false
	}

	if err := bcrypt.CompareHashAndPassword([]byte(userInfo.Password), []byte(password)); err == nil {
		return userInfo.Account, true
	} else {
		logrus.Debug(err)
		return userInfo.Account, false
	}
}

// Me 当前账户信息
func Me(c *gin.Context) {
	a := jwt.ExtractClaims(c)["id"].(string)

	user, err := model.FindUserByAccount(a)
	if !handleError(err, c) {
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"account":          user.Account,
		"password_changed": user.PasswordChanged,
	})

}

type UpdatePasswordReq struct {
	NewPassword string `j:"new"`
}

func UpdatePassword(c *gin.Context) {
	a := jwt.ExtractClaims(c)["id"].(string)

	user, err := model.FindUserByAccount(a)
	if !handleError(err, c) {
		return
	}

	var newPwd UpdatePasswordReq
	err = c.BindJSON(&newPwd)
	if !handleError(err, c) {
		return
	}
	if newPwd.NewPassword == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"dormon":  "",
			"message": "params [new] is required",
		})
		return
	}

	// TODO:实际上传过来的是MD5，长度不会小于8
	if strings.Count(newPwd.NewPassword, "") < 8 {
		c.JSON(http.StatusBadRequest, gin.H{
			"dormon":  "",
			"message": "new password is toooooooooo short",
		})
		return
	}

	password, _ := bcrypt.GenerateFromPassword(
		[]byte(newPwd.NewPassword),
		bcrypt.DefaultCost,
	)

	user.Password = string(password)
	err = user.UpdatePassword()
	if !handleError(err, c) {
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"dormon":  "",
		"message": "ok",
	})
}
