package i18n

import (
	"fmt"
	"github.com/magiconair/properties"
	"log"
)

var P *properties.Properties

const (
	myLanguage      = "russian"
	defaultLanguage = "english"
)

func init() {
	var fileNames []string
	for _, l := range []string{defaultLanguage, myLanguage} {
		fileNames = append(fileNames, fmt.Sprintf("i18n/%s.properties", l))
	}

	var err error
	P, err = properties.LoadFiles(fileNames, properties.UTF8, false)
	if err != nil {
		log.Fatalf("can't open file fo i18n: %s", err)
	}
}
