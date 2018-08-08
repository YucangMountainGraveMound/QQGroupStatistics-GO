package utils

import (
	"os"
	"net/http"
	"io/ioutil"
	"io"
	"bytes"

	"encoding/base64"
	"dormon.net/qq/errors"
	"github.com/sirupsen/logrus"
)

// PathExists check the file existence or create it while create is true
func PathExists(path string, create bool) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		if create {
			err := os.Mkdir(path, os.ModePerm)
			return err == nil, err
		} else {
			return false, nil
		}
	}
	return false, err
}

func DownloadImageToBase64(imageUrl string) (string, error) {
	resp, err := http.Get("http://42.202.154.18" + imageUrl)
	if err != nil {
		return "", err
	}

	if resp.StatusCode != 200 {
		logrus.Errorf("Failed to download file with status code of %d", resp.StatusCode)
		return "", errors.New("error get")
	}

	contentType := resp.Header.Get("Content-Type")

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	// body to base64

	picBase64 := base64.StdEncoding.EncodeToString(body)
	hash := MD5(picBase64)

	var name string
	path := "./images/"

	switch contentType {
	case "image/jpeg":
		name = hash + ".jpg"
		break
	case "image/png":
		name += hash + ".png"
		break
	case "image/gif":
		name += hash + ".gif"
		break
	default:
		name = RandomString(16) + ".unknown"
	}

	out, err := os.Create(path + name)
	if err != nil {
		return "", err
	}
	_, err = io.Copy(out, bytes.NewReader(body))
	if err != nil {
		return "", err
	}
	return name, nil
}
