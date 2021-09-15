package i18n

import (
	"fmt"
	"github.com/BurntSushi/toml"
	"github.com/SongOf/edge-storage-core/pkg/eslog"
	"github.com/nicksnyder/go-i18n/v2/i18n"
	"golang.org/x/text/language"
	"io/ioutil"
	"path/filepath"
	"regexp"
)

func checkFilename(filename string) bool {
	regObj, err := regexp.Compile(`^active\..*?\.toml$`)
	if err != nil {
		eslog.L().Panic("compile regobj error", eslog.Err(err))
	}
	return regObj.MatchString(filename)
}

func NewTranslator(msgFileDir string) (*Translator, error) {

	bundle := i18n.NewBundle(language.English)
	bundle.RegisterUnmarshalFunc("toml", toml.Unmarshal)
	files, err := ioutil.ReadDir(msgFileDir)
	if err != nil {
		return nil, err
	}

	for _, file := range files {
		if !checkFilename(file.Name()) {
			return nil, fmt.Errorf("the name of message file `%s` is invalid", file.Name())
		}

		filePath := filepath.Join(msgFileDir, file.Name())
		_, err := bundle.LoadMessageFile(filePath)
		if err != nil {
			return nil, fmt.Errorf("load message file:%v", err)
		}
	}

	return &Translator{
		bundle: bundle,
	}, nil
}

type Translator struct {
	bundle      *i18n.Bundle
	fileNameReg *regexp.Regexp
}

func (t *Translator) Translate(lang, code, defaultMsg string, data interface{}) (string, error) {
	localizer := i18n.NewLocalizer(t.bundle, lang)
	msg, err := localizer.Localize(&i18n.LocalizeConfig{
		TemplateData: data,
		DefaultMessage: &i18n.Message{
			ID:    code,
			Other: defaultMsg,
		},
	})
	if err != nil {
		if _, ok := err.(*i18n.MessageNotFoundErr); !ok {
			return "", nil
		}
	}
	return msg, nil
}
