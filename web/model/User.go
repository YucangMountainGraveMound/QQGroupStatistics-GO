package model

import (
	"dormon.net/qq/config"
	"dormon.net/qq/utils"

	"github.com/jinzhu/gorm"
	"github.com/sirupsen/logrus"
	"golang.org/x/crypto/bcrypt"
	"dormon.net/qq/db"
)

type User struct {
	gorm.Model
	Account         string `gorm:"type:varchar(20);unique_index"`
	Password        string `gorm:"type:varchar(100)"`
	PasswordChanged bool   `gorm:"type:bool;default:'false'"`
}

func CreateUser(user *User) error {
	return db.GetDB().Where(User{Account: user.Account}).FirstOrCreate(&user).Error
}

// FindUserByAccount 根据用户account查找用户
func FindUserByAccount(account string) (*User, error) {
	var user User

	err := db.GetDB().First(&user, "account = ?", account).Error

	return &user, err
}

// UpdatePassword 修改密码
func (user *User) UpdatePassword() error {
	return db.GetDB().Model(&user).Updates(map[string]interface{}{
		"password":         user.Password,
		"password_changed": true,
	}).Error
}

// 初始化用户账号
// 约定初始账号为qq号码，密码为 “qq” + qq号码
// 由于聊天记录里都是昵称或者备注，相关联信息在user_alias表中
func InitialAccount() {
	for _, value := range config.Config().Account {

		password, _ := bcrypt.GenerateFromPassword(
			[]byte(utils.MD5(utils.MD5("qq"+value.Account)+value.Account)),
			bcrypt.DefaultCost,
		)
		user := User{
			Account:         value.Account,
			Password:        string(password),
			PasswordChanged: false,
		}

		if err := CreateUser(&user); err != nil {
			logrus.Fatalf("Error when creating user %s: %s", user, err)
		}
	}
}
