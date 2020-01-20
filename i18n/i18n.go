package i18n

import (
	"fmt"
	"github.com/magiconair/properties"
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

	P = properties.MustLoadFiles(fileNames, properties.UTF8, false)
}
