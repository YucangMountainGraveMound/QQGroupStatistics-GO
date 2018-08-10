package record_process

import (
	"bufio"
	"bytes"
	"encoding/base64"
	"fmt"
	"io"
	"io/ioutil"
	"mime"
	"net/textproto"
	"strings"

	"dormon.net/qq/record_process/multipart"
	"dormon.net/qq/utils"
	"dormon.net/qq/web/model"

	"github.com/PuerkitoBio/goquery"
	"gopkg.in/cheggaaa/pb.v1"
)

var _ = ioutil.ReadAll

type MHtml struct {
	Html  string
	Image map[string]Image
}

type Image struct {
	Type     string
	Location string
	Content  string
}

func New() *MHtml {
	return &MHtml{}
}

// Parse 解析导出的mht格式聊天记录
func (m *MHtml) Parse(mht []byte) error {
	if err := m.ParseContent(mht); err != nil {
		return err
	}

	parseResources(m)
	ParseHtml(m)
	return nil
}

func (m *MHtml) ParseContent(mht []byte) error {

	br := bufio.NewReader(bytes.NewReader(mht))
	tr := textproto.NewReader(br)
	boundary := m.GetBoundary(tr)
	if len(boundary) == 0 {
		return fmt.Errorf("获取边界失败")
	}

	mr := multipart.NewReader(br, boundary)
	image := Image{}
	m.Image = make(map[string]Image)
	for {
		d := make([]byte, 10*1024*1024)

		part, err := mr.NextPart()
		if err != nil {
			break
		}

		n, err := part.Read(d)
		if err != nil && err != io.EOF {
			return err
		}
		d = d[:n]
		if len(part.Header["Content-Type"]) == 0 {
			continue
		}
		if part.Header["Content-Type"][0] == "text/html" {
			m.Html = string(d)
		} else {
			image.Type = part.Header["Content-Type"][0]
			image.Location = part.Header["Content-Location"][0]
			image.Content = string(d)
			m.Image[utils.MD5(string(d))] = image
		}
	}
	return nil
}

func (m *MHtml) GetBoundary(r *textproto.Reader) string {
	mimeHeader, err := r.ReadMIMEHeader()
	if err != nil {
		return ""
	}
	contentType := mimeHeader.Get("Content-Type")

	_, params, _ := mime.ParseMediaType(contentType)
	boundary := params["boundary"]
	return boundary
}

func ParseHtml(m *MHtml) error {
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(m.Html))
	if err != nil {
		panic(err)
	}

	total := doc.Find("table").Find("tr").Find("td").Length()

	bar := pb.StartNew(int(total))

	date := ""
	doc.Find("table").Find("tr").Find("td").Each(func(i int, selection *goquery.Selection) {
		if selection.Children().Size() == 0 {
			date = selection.Text()
		}

		if selection.Children().Size() == 2 {
			// 用户账号
			account := selection.Get(0).FirstChild.FirstChild.FirstChild.Data

			// 时间
			time := selection.Get(0).FirstChild.FirstChild.NextSibling.Data
			formattedTime := utils.FormatDateTime(date, time)

			// 消息文字内容
			message := ""
			at := ""
			selection.Children().Find("font").Each(func(i int, selection *goquery.Selection) {
				if len(selection.Text()) != 0 {
					if selection.Text()[0:1] == "@" {
						// @了谁
						at = selection.Text()[1:]
					} else {
						message += selection.Text()
					}
				}
			})

			// 消息图片内容
			images := make([]string, selection.Find("IMG").Length())
			imageIndex := 0
			selection.Find("IMG").Each(func(i int, selection *goquery.Selection) {
				// 由于导出的html img src 中可能存在空格问题
				image, _ := selection.Attr("src")
				if image != "" {
					imageName, _ := model.FindImageByImageHash(image)
					images[imageIndex] = imageName.ImageName
					imageIndex ++
				}
			})

			model.CreateRecordFromImport(&model.Record{
				Number:     account,
				Message:    message,
				Date:       formattedTime,
				Images:     images[:],
				At:         at,
				MessageLen: strings.Count(message, "") - 1,
			})

		}
		bar.Increment()
	})

	bar.FinishPrint("Done")
	return nil
}

func parseResources(m *MHtml) error {

	var err error
	var dbErr error

	_, err = utils.PathExists("./images", true)
	if err != nil {
		panic(err)
	}

	for k, v := range m.Image {

		switch v.Type {
		case "image/jpeg":
			image, _ := base64.StdEncoding.DecodeString(v.Content)
			err = ioutil.WriteFile("./images/"+k+".jpg", image, 0666)
			dbErr = model.CreateImage(k+".jpg", v.Location)
			break
		case "image/png":
			image, _ := base64.StdEncoding.DecodeString(v.Content)
			err = ioutil.WriteFile("./images/"+k+".png", image, 0666)
			dbErr = model.CreateImage(k+".png", v.Location)
			break
		case "image/gif":
			image, _ := base64.StdEncoding.DecodeString(v.Content)
			err = ioutil.WriteFile("./images/"+k+".gif", image, 0666)
			dbErr = model.CreateImage(k+".gif", v.Location)
			break
		}
		if err != nil || dbErr != nil {
			panic(err)
		}
	}
	return nil
}
