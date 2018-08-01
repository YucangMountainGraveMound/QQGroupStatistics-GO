package model

import (
	"dormon.net/qq/db"

	"github.com/jinzhu/gorm"
)

type Image struct {
	gorm.Model
	ImageName string `gorm:"type:varchar(200)"`
	ImageHash string `gorm:"type:varchar(200)"`
}

func CreateImage(imageName, imageHash string) error {
	image := &Image{
		ImageName: imageName,
		ImageHash: imageHash,
	}
	return db.GetDB().FirstOrCreate(image, "image_name = ? and image_hash = ?", imageName, imageHash).Error
}

func FindImageByImageHash(imageHash string) (*Image, error) {
	var image Image

	err := db.GetDB().First(&image, "image_hash = ?", imageHash).Error

	return &image, err
}
