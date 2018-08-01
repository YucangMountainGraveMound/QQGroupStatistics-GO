package utils

import (
	"os"
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
