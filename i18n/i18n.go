package i18n

import (
	"fmt"
	"github.com/magiconair/properties"
	"log"
)

var P labels

const (
	myLanguage      = "russian"
	defaultLanguage = "english"
)

func Init() {
	var fileNames []string
	for _, l := range []string{defaultLanguage, myLanguage} {
		fileNames = append(fileNames, fmt.Sprintf("i18n/%s.properties", l))
	}

	var err error
	props, err := properties.LoadFiles(fileNames, properties.UTF8, false)
	if err != nil {
		log.Fatalf("can't open file fo i18n: %s", err)
	}
	P = labelsProp{props}
}

func TestInit() {
	P = labelsMock{}
}

type labels interface {
	MustGet(string) string
	Get(string) (string, bool)
	Keys() []string
}

type labelsProp struct {
	P *properties.Properties
}

func (l labelsProp) Keys() []string {
	return l.P.Keys()
}

func (l labelsProp) Get(lbl string) (string, bool) {
	return l.P.Get(lbl)
}

func (l labelsProp) MustGet(lbl string) string {
	return l.P.MustGetString(lbl)
}

type labelsMock struct {
}

func (l labelsMock) MustGet(string) string {
	return ""
}

func (l labelsMock) Get(string) (string, bool) {
	panic("implement me labelsMock Get")
}

func (l labelsMock) Keys() []string {
	return []string{}
}
