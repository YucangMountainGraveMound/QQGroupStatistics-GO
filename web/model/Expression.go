package model

import "dormon.net/qq/db"

type Expression struct {
	ID   string `gorm:"primary_key"`
	Name string `gorm:"type:varchar(10)"`
	Hash string `gorm:"type:varchar(100)"`
}

func getExpressionByName(name string) (Expression, error) {
	var exp Expression
	err := db.GetDB().First(&exp, "name = ?", name).Error
	return exp, err
}

func getExpressionByHash(hash string) (Expression, error) {
	var exp Expression
	err := db.GetDB().First(&exp, "hash = ?", hash).Error
	return exp, err
}