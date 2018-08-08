package record_process

import (
	"dormon.net/qq/web/model"
	"dormon.net/qq/utils"
	"github.com/sirupsen/logrus"
	"strings"
)

var MsgChan = make(chan model.Message, 100)
var PicChan = make(chan model.Picture, 100)

func Process() {

	for msg := range MsgChan {
		if msg.ClassName == "MessageForPic" || msg.ClassName == "MessageForMixedMsg" {
			var picArr []string
			for i := 0; i < len(PicChan); i++ {
				pic := <-PicChan
				if msg.UniSeq == pic.UniSeq {
					hash, err := utils.DownloadImageToBase64(pic.PicUrl)
					if err != nil {
						logrus.Errorf("Failed to download image with error: %s", err)
					}

					picArr = append(picArr, hash)

					if len(picArr) == strings.Count(msg.Message, "[图片]") {
						break
					}
				} else {
					PicChan <- pic
				}
			}

			if len(picArr) != strings.Count(msg.Message, "[图片]") {
				MsgChan <- msg
				logrus.Info("No or not enough match pic found, recycle")
			} else {
				logrus.Info(picArr)
			}

		} else {
			logrus.Info(msg)
		}
	}

	logrus.Info("exit?")
}
