package log

import "github.com/sirupsen/logrus"

var (
	Log = logrus.New()
)

func LogErrorFrom(method string, errFrom string, err error){
	Log.WithFields(logrus.Fields{
		"method": method,
		"errFrom": errFrom,
	}).Error(err)
}