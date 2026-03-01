package i18n

import "testing"

func TestAllKeysHaveTranslations(t *testing.T) {
	en, err := LoadTranslations("en")
	if err != nil {
		t.Fatalf("failed to load EN: %v", err)
	}
	ru, err := LoadTranslations("ru")
	if err != nil {
		t.Fatalf("failed to load RU: %v", err)
	}

	for key := range en {
		if _, ok := ru[key]; !ok {
			t.Errorf("key %q exists in EN but missing in RU", key)
		}
	}
	for key := range ru {
		if _, ok := en[key]; !ok {
			t.Errorf("key %q exists in RU but missing in EN", key)
		}
	}
}

func TestLoadTranslationsInvalidLang(t *testing.T) {
	_, err := LoadTranslations("xx")
	if err == nil {
		t.Error("expected error for unknown language, got nil")
	}
}

func TestTr(t *testing.T) {
	SetLang("en")
	if got := Tr("dashboard.no_focus"); got == "dashboard.no_focus" {
		t.Error("expected translated string, got key back")
	}
	if got := Tr("nonexistent.key"); got != "nonexistent.key" {
		t.Errorf("expected key returned for missing translation, got %q", got)
	}
}
