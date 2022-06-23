package utils

import "github.com/sirupsen/logrus"

func Panic(err error) {
	if err != nil {
		logrus.Fatal(err)
	}
}
