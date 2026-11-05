package morse_test

import (
	"testing"

	"github.com/Yandex-Practicum/go1fl-sprint6-final/pkg/morse"
)

func TestWordSeparators(t *testing.T) {
	tests := []struct {
		name, charSeparator, wordSeparator, want string
	}{
		{"default", " ", "   ", ".- -...   .-- --."},
		{"custom", " | ", " / ", ".- | -... / .-- | --."},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			converter := morse.NewConverter(morse.DefaultMorse,
				morse.WithCharSeparator(tt.charSeparator),
				morse.WithWordSeparator(tt.wordSeparator))
			encoded := converter.ToMorse("АБ ВГ")
			if encoded != tt.want {
				t.Fatalf("encoded=%q; want %q", encoded, tt.want)
			}
			if decoded := converter.ToText(encoded); decoded != "АБ ВГ" {
				t.Fatalf("decoded=%q; want %q", decoded, "АБ ВГ")
			}
		})
	}
}
