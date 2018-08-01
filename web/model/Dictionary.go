package model

import (
	"dormon.net/qq/db"

	"github.com/jinzhu/gorm"
)

type Dictionary struct {
	gorm.Model
	Key   string `gorm:"type:varchar(100);unique_index"`
	Value string `gorm:"type:text"` // TODO:使用postgres原生的数组存取
}

func (dict *Dictionary) FirstOrCreate() (err error) {
	return db.GetDB().FirstOrCreate(dict, "key = ?", dict.Key).Error
}

func (dict *Dictionary) Update() (err error) {
	return db.GetDB().Model(&dict).Updates(map[string]interface{}{
		"value": dict.Value,
	}).Error
}

func GetDictionaryByKey(key string) (Dictionary, error) {
	var dict Dictionary

	err := db.GetDB().First(&dict, "key = ?", key).Error

	return dict, err
}
