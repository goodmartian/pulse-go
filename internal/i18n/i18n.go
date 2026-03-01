package i18n

import (
	"embed"
	"encoding/json"
	"fmt"

	"github.com/goodmartian/pulse-go/internal/config"
)

//go:embed locales/*.json
var localesFS embed.FS

var (
	Lang string
	t    map[string]string
)

func LoadTranslations(lang string) (map[string]string, error) {
	data, err := localesFS.ReadFile("locales/" + lang + ".json")
	if err != nil {
		return nil, fmt.Errorf("load locale %s: %w", lang, err)
	}
	var m map[string]string
	if err := json.Unmarshal(data, &m); err != nil {
		return nil, fmt.Errorf("parse locale %s: %w", lang, err)
	}
	return m, nil
}

func Tr(key string, args ...any) string {
	s, ok := t[key]
	if !ok {
		return key
	}
	if len(args) > 0 {
		return fmt.Sprintf(s, args...)
	}
	return s
}

func Init() {
	cfg := config.LoadConfig()
	Lang = cfg.Lang
	if Lang == "" {
		state := config.LoadState()
		if state.Lang != "" {
			Lang = state.Lang
		} else {
			Lang = "en"
		}
	}
	SetLang(Lang)
}

func SetLang(lang string) {
	translations, err := LoadTranslations(lang)
	if err != nil {
		Lang = "en"
		translations, _ = LoadTranslations("en")
	}
	Lang = lang
	t = translations
}
