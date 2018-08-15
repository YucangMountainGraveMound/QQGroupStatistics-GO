package es

import (
	"context"
	"dormon.net/qq/config"
	"io"
	"encoding/json"
	"github.com/sirupsen/logrus"
	"gopkg.in/cheggaaa/pb.v1"
	"time"
	"strings"
	"gopkg.in/urfave/cli.v1"
)

var CMDRunTrim = cli.Command{
	Name:        "trim",
	Usage:       "trim data",
	Description: "Analyze all data and do some trim jobs",
	Flags: []cli.Flag{
		cli.StringFlag{
			EnvVar: "DORMON_QQ_CONFIG",
			Name:   "config, C",
			Usage:  "specify the configuration file",
			Value:  "./config.toml",
		},
		cli.BoolFlag{
			EnvVar: "DORMON_QQ_GENERATE_CONFIG",
			Name:   "generate, G",
			Usage:  "generate a configuration file",
		},
	},
	Action: TrimESDate,
}

type Record struct {
	Number      string    `json:"number"`
	Message     string    `json:"message"`
	Date        time.Time `json:"date"`
	Images      []string  `json:"images"`
	At          string    `json:"at"`
	Interval    int64     `json:"interval"`
	MessageLen  int       `json:"message_len"`
	Expression  []string  `json:"expression"`
	MessageType string    `json:"message_type"`
}

func TrimESDate(c *cli.Context) error {

	config.InitialConfig(c)

	esClient := ElasticClient()

	index := config.Config().ElasticSearchConfig.AliasName

	scroll := esClient.Scroll(index).Type(index).Sort("date", true).Size(100)

	temp := map[string]time.Time{}

	total, err := esClient.Count(index).Type(index).Do(context.Background())
	if err != nil {
		panic(err)
	}
	bar := pb.StartNew(int(total))

	for {
		results, err := scroll.Do(context.Background())
		if err == io.EOF {
			return nil
		}
		if err != nil {
			return err
		}

		for _, hit := range results.Hits.Hits {
			bar.Increment()

			var record Record
			err := json.Unmarshal(*hit.Source, &record)
			if err != nil {
				logrus.Errorf("Failed to decode json to record with error: %s", err)
			}

			// 删除无效数据
			if record.Number == "" || record.Number == "10000" {
				//logrus.Infof("deleting doc %s", hit.Id)
				_, err = esClient.Delete().Index(index).Type(index).Id(hit.Id).Do(context.Background())
				if err != nil {
					logrus.Errorf("Trim failed to delete illegal doc of %s with error: %s", hit.Id, err)
				}
				continue
			}

			// 处理空图片和表情
			if len(record.Images) != 0 {
				var newImages []string
				exp := record.Expression
				for _, image := range record.Images {
					if image != "" {
						i := findExpFromLocalImages(image)
						if i != "" {
							exp = append(exp, i)
						} else {
							newImages = append(newImages, image)
						}
					}
				}
				if len(record.Images) == 0 {
					newImages = nil
				}
				record.Expression = exp
				record.Images = newImages
			}

			// 过滤相同的图片

			// 处理interval
			if temp[record.Number].IsZero() {
				record.Interval = -1
				temp[record.Number] = record.Date
			} else {
				record.Interval = record.Date.Unix() - temp[record.Number].Unix()
				temp[record.Number] = record.Date
			}

			// 处理消息内容
			strings.Replace(record.Message, "[图片]", "", -1)

			// 处理消息长度
			record.MessageLen = strings.Count(record.Message, "") - 1

			// 处理消息类型
			if record.MessageLen == 0 && len(record.Images) == 0 && len(record.Expression) == 0 {
				// 非法消息，删除
				_, err = esClient.Delete().Index(index).Type(index).Id(hit.Id).Do(context.Background())
				if err != nil {
					logrus.Errorf("Trim failed to delete illegal doc of %s with error: %s", hit.Id, err)
				}
				continue
			}
			if record.MessageLen != 0 || len(record.Expression) != 0 {
				if len(record.Images) != 0 {
					record.MessageType = "MessageForMixedMsg"
				} else {
					record.MessageType = "MessageForText"
				}
			} else {
				record.MessageType = "MessageForPic"
			}

			_, err = esClient.Update().Index(index).Type(index).Id(hit.Id).Doc(record).Do(context.Background())
			if err != nil {
				logrus.Errorf("Trim failed to update doc of %s with error: %s", hit.Id, err)
			}
		}
	}

	bar.FinishPrint("Done")
	return nil
}

func findExpFromLocalImages(image string) string {
	var exp string
	switch image {
	case "53f3a9b5d6887bccf8cce2735e489562.png":
		exp = "57.png"
		break
	case "b8029ee8591d993bac65bc839201fae9.gif":
		exp = "64.png"
		break
	case "2210684100367bbc6bfba92c3202ce55.gif":
		exp = "58.png"
		break
	case "d687ddedcb0d141ee5a94cbc5088d4bd.gif":
		exp = "63.png"
		break
	case "852fb35da8ede2bb760e09d223c7bc50.gif":
		exp = "1.png"
		break
	case "a10da2827da3614ecdd652ed42b3677d.gif":
		exp = "14.png"
		break
	case "ee92c08cd2b72b10b51996c31cf78e12.gif":
		exp = "21.png"
		break
	case "92f1afb678b87929f1ff994713b854d5.gif":
		exp = "59.png"
		break
	case "5843a2ba75819270c5335a42c7a315ae.gif":
		exp = "24.png"
		break
	case "f53a4c744ac33c70dc90962cd27ffca9.gif":
		exp = "10.png"
		break
	case "dd2af13bc060694bcac83234cdb7492a.gif":
		exp = "67.png"
		break
	case "6bbb852495d5c376041e4a8375f3bef5.gif":
		exp = "17.png"
		break
	case "c56abd98c150576d823e91432892a4ae.gif":
		exp = "40.png"
		break
	case "5d2d7e2cb465f406600270ce2bca2584.gif":
		exp = "51.png"
		break
	case "ed1be07ecd0b7423afec659c76cd01cc.gif":
		exp = "5.png"
		break
	case "7b00b19b11b9d7a3e59a374bad81974b.gif":
		exp = "45.png"
		break
	case "a3635b96571479fb2d2f37aa6db178d1.gif":
		exp = "61.png"
		break
	case "d81f3034ffd3b8ea564b9bfa48d38e25.gif":
		exp = "57.png"
		break
	case "7a0cf162f95cee65116a1ad7e1f25d60.gif":
		exp = "65.png"
		break
	case "0b942ecacfe889a576d31585dce5a5ee.gif":
		exp = "49.png"
		break
	case "8d33cd445a4e5cbe4e5938af21961207.gif":
		exp = "2.png"
		break
	case "a5535d36ed30652b204cd8b5e0474ab5.gif":
		exp = "62.png"
		break
	case "1696b6ddb8abe545e2ae0fc86be7369d.gif":
		exp = "3.png"
		break
	case "88ed0b08f2a35c6a399f0d4e0bb7d44d.gif":
		exp = "9.png"
		break
	case "14f0a0597c9430c0fdc39806f37b139f.gif":
		exp = "11.png"
		break
	case "756154834bed7bd9477ed568af1dd9be.gif":
		exp = "18.png"
		break
	case "c662ad7faf3e48494fc7b1c3e5cbf952.gif":
		exp = "4.png"
		break
	case "5654f0f862a31213a1ef9c11ceda9dfd.gif":
		exp = "25.png"
		break
	case "a5853326005c7fea790c8c3cc4ab123b.gif":
		exp = "26.png"
		break
	case "bae14ea5e3016744f82bd991ed561467.gif":
		exp = "35.png"
		break
	case "af783f921473a18a33166f8865dfda1b.gif":
		exp = "52.png"
		break
	case "3e4514c4c279ba7ab49ac0a9f5369db0.gif":
		exp = "7.png"
		break
	case "02237bc2d6706bd6a1fa89e11d4f4386.gif":
		exp = "23.png"
		break
	case "7f144c0383978a05ea497fae2eda4d11.gif":
		exp = "29.png"
		break
	case "e9d38caab77d9dc53f216ae744c540ca.gif":
		exp = "10.png"
		break
	case "eb5008735a42af9a69f04d5f99ba8c63.gif":
		exp = "28.png"
		break
	case "38ff7890dce55246c4a5bc3f8ccc6661.gif":
		exp = "60.png"
		break
	case "e7af53e596bb7bf285bd0198a0a79fd1.gif":
		exp = "27.png"
		break
	case "633928d36140303b4151c38bdfbdd088.gif":
		exp = "8.png"
		break
	case "8d52395940e078f94b2595a39bdaf0e9.gif":
		exp = "32.png"
		break
	case "66fe5533656fd33e49819aef3f1dc656.gif":
		exp = "33.png"
		break
	case "76cee1f0104cdcf8b03c7fec3b010d82.gif":
		exp = "44.png"
		break
	case "13fd44da413309e4b106543247268d2b.gif":
		exp = "15.png"
		break
	case "5373a37e9fd546518173b93ba0ff0add.gif":
		exp = "12.png"
		break
	case "7e1f63ca676187be874fc4f0236745c9.gif":
		exp = "66.png"
		break
	case "b1f21ae4173b94c8dbf471bfc2950bea.gif":
		exp = "42.png"
		break
	case "c07e6f6e0ff14fc99b0fe9298e302463.gif":
		exp = "42.png"
		break
	case "53d0559982d59c1865247f5df0dcd054.gif":
		exp = "6.png"
		break
	case "c8ba815852f253c25c06cce0fbb28782.gif":
		exp = "48.png"
		break
	case "765e9c0ce5c464f6a7620b224a3b1293.gif":
		exp = "13.png"
		break
	case "e225a7e3cba674d210afa0af2c83c8fc.gif":
		exp = "20.png"
		break
	case "0e76cc2a4d546c4f861c944b8410e316.gif":
		exp = "50.png"
		break
	case "f2f08f49b77b3eb8ee18ecd3925100bd.gif":
		exp = "43.png"
		break
	case "0165d891141a57894452b316d0e3b601.gif":
		exp = "41.png"
		break
	case "5abed65813bf144cb3da56530351b723.gif":
		exp = "73.gif"
		break
	case "2840010e2931d40e7ac30598a37a3b65.gif":
		exp = "74.gif"
		break
	case "2b8edefe39e8aac996a9856ab469ca35.gif":
		exp = "75.gif"
		break
	case "57e0afd92d103ac8c0558445ad399cd9.gif":
		exp = "37.png"
		break
	case "2e7414bceada27f5ff863c098f3d929f.gif":
		exp = "76.gif"
		break
	case "76e74768c81ed0e143dce29037ab438b.png":
		exp = "77.png"
		break
	}
	return exp
}
