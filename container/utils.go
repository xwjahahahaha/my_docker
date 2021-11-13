package container

import (
	"os"
	"xwj/mydocker/log"
)

func dirOrFileExist(path string) (bool, error) {
	if _, err := os.Stat(path); err == nil {
		return true, nil
	}else if os.IsNotExist(err) {
		return false, nil
	}else if os.IsNotExist(err) {
		return false, nil
	}else {
		log.Log.Error(err)
		return false, err
	}
}
