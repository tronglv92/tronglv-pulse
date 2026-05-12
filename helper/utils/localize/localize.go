package localize

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/nicksnyder/go-i18n/v2/i18n"
	"github.com/zeromicro/go-zero/core/logx"
	"golang.org/x/text/language"
	"sync"
)

type ContextKey string

const (
	LanguageKey ContextKey = "language"
	LocalizeKey ContextKey = "localize"
	Fallback               = "vi"
)

type (
	Localizer interface {
		Get(key string) (string, error)
		GetString(key string) string
		GetFormat(key string, v ...any) string
		GetFromLocale(locale string, key string) string
		GetFromContext(ctx context.Context, key string) string
		SetLocale(locale string) Localizer
	}

	Config interface {
		GetLanguage() string
	}

	localizeService struct {
		defaultLang string
		bundle      *i18n.Bundle
		localizer   *i18n.Localizer
		locales     map[string]string
		once        sync.Once
	}

	Language struct {
		Lang string
	}
)

func (l Language) GetLanguage() string {
	return l.Lang
}

func NewWithDefault() Localizer {
	return NewLocalizeService(
		Language{Lang: language.English.String()},
		Language{Lang: language.Vietnamese.String()},
	)
}

func NewLocalizeService(configs ...Config) Localizer {
	s := &localizeService{
		defaultLang: language.Vietnamese.String(),
		locales:     make(map[string]string),
	}
	s.once.Do(func() {
		s.bundle = i18n.NewBundle(language.English)
		s.bundle.RegisterUnmarshalFunc("json", json.Unmarshal)
		for _, c := range configs {
			s.locales[c.GetLanguage()] = c.GetLanguage()
			if err := s.LoadMessageFile(c.GetLanguage()); err != nil {
				logx.Error(err)
			}
		}
		s.localizer = i18n.NewLocalizer(s.bundle, s.defaultLang)
	})
	return s
}

func (s *localizeService) Get(key string) (string, error) {
	return s.localize(key)
}

func (s *localizeService) GetString(key string) string {
	msg, err := s.localize(key)
	if err != nil {
		logx.Infof("localization failed for key %s: %v", key, err)
		return ""
	}
	return msg
}

func (s *localizeService) GetFromContext(ctx context.Context, key string) string {
	return s.SetLocale(s.getLocaleFromContext(ctx)).GetString(key)
}

func (s *localizeService) GetFromLocale(locale string, key string) string {
	return s.SetLocale(locale).GetString(key)
}

func (s *localizeService) GetFormat(key string, v ...any) string {
	return fmt.Sprintf(s.GetString(key), v...)
}

func (s *localizeService) getLocaleFromContext(ctx context.Context) string {
	if locale, ok := ctx.Value(LanguageKey).(string); ok && s.validLocale(locale) {
		return locale
	}
	return Fallback
}

func (s *localizeService) LoadMessageFile(lang string) error {
	if !s.validLocale(lang) {
		return fmt.Errorf("localize %s does not exist", lang)
	}
	_, err := s.bundle.LoadMessageFile(fmt.Sprintf("locale/%s.json", lang))
	return err
}

func (s *localizeService) validLocale(locale string) bool {
	_, exists := s.locales[locale]
	return exists
}

func (s *localizeService) Bundle() *i18n.Bundle {
	return s.bundle
}

func (s *localizeService) SetLocale(locale string) Localizer {
	return &localizeService{
		bundle:    s.Bundle(),
		localizer: i18n.NewLocalizer(s.bundle, locale),
	}
}

func (s *localizeService) localize(key string) (string, error) {
	return s.localizer.Localize(&i18n.LocalizeConfig{MessageID: key})
}
