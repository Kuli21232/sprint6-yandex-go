package service_test

import (
	"testing"

	"github.com/Yandex-Practicum/go1fl-sprint6-final/internal/service"
)

func TestConvertSupportedInputs(t *testing.T) {
	tests := []struct{ name, input, want string }{
		{"russian_lowercase", "привет", ".--. .-. .. .-- . -"},
		{"russian_uppercase", "ПРИВЕТ", ".--. .-. .. .-- . -"},
		{"morse", ".--. .-. .. .-- . -", "ПРИВЕТ"},
		{"outer_whitespace", "\r\n .--. .-. .. .-- . - \t", "ПРИВЕТ"},
		{"single_letter", "я", ".-.-"},
		{"single_signal", ".", "Е"},
		{"digits", "0123456789", "----- .---- ..--- ...-- ....- ..... -.... --... ---.. ----."},
		{"digits_back", "----- .---- ..--- ...-- ....- ..... -.... --... ---.. ----.", "0123456789"},
		{"punctuation", "А,Б?", ".- .-.-.- -... ..--.."},
		{"morse_word_separator", ".-   -...", "А Б"},
		{"text_lf", "А\nБ", ".-   -..."},
		{"text_unicode_whitespace", "А\u00a0Б", ".-   -..."},
		{"morse_unicode_whitespace", "...\u00a0---\u00a0...", "СОС"},
		{"repeated_spaces", "А  Б", ".-      -..."},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := service.Convert(tt.input)
			if err != nil || got != tt.want {
				t.Fatalf("Convert(%q) = %q, %v; want %q, nil", tt.input, got, err, tt.want)
			}
		})
	}
}

func TestConvertEmptyInputs(t *testing.T) {
	for _, input := range []string{"", " ", "\r\n\t", "\u00a0\u2003"} {
		if got, err := service.Convert(input); err == nil {
			t.Errorf("Convert(%q) = %q, nil; want an error", input, got)
		}
	}
}

func TestConvertEdgeCases(t *testing.T) {
	tests := []struct {
		name, input, want string
		wantErr           bool
	}{
		{name: "preserve_word_boundary", input: "А Б", want: ".-   -..."},
		{name: "morse_lf", input: "...\n---\n...", want: "СОС"},
		{name: "morse_crlf", input: "...\r\n---\r\n...", want: "СОС"},
		{name: "morse_tab", input: "...\t---\t...", want: "СОС"},
		{name: "morse_utf8_bom", input: "\ufeff... --- ...", want: "СОС"},
		{name: "unsupported_only", input: "🙂", wantErr: true},
		{name: "unsupported_mixed", input: "А🙂Б", wantErr: true},
		{name: "unsupported_latin", input: "Hello", wantErr: true},
		{name: "invalid_utf8", input: string([]byte{0xff, 0xfe}), wantErr: true},
		{name: "unknown_morse_token", input: ".......", wantErr: true},
		{name: "unknown_morse_token_mixed", input: "... ....... ---", wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := service.Convert(tt.input)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("Convert(%q) = %q, nil; want an explicit error instead of silent data loss", tt.input, got)
				}
				return
			}
			if err != nil || got != tt.want {
				t.Fatalf("Convert(%q) = %q, %v; want %q, nil", tt.input, got, err, tt.want)
			}
		})
	}
}

func TestConvertRoundTrip(t *testing.T) {
	for _, text := range []string{"А Б", "А  Б", "ПРИВЕТ МИР", "А, Б? 123"} {
		encoded, err := service.Convert(text)
		if err != nil {
			t.Fatal(err)
		}
		decoded, err := service.Convert(encoded)
		if err != nil || decoded != text {
			t.Errorf("round trip %q -> %q -> %q, err=%v", text, encoded, decoded, err)
		}
	}
}
