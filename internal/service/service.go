package service

import (
	"errors"
	"fmt"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/Yandex-Practicum/go1fl-sprint6-final/pkg/morse"
)

func IsMorse(s string) bool {
	s = strings.TrimPrefix(s, "\ufeff")
	if len(strings.TrimSpace(s)) == 0 {
		return false
	}
	for _, ch := range s {
		if ch != '.' && ch != '-' && !unicode.IsSpace(ch) {
			return false
		}
	}
	return true
}

func Convert(s string) (string, error) {
	if !utf8.ValidString(s) {
		return "", errors.New("файл должен содержать корректный UTF-8")
	}
	s = strings.TrimSpace(strings.TrimPrefix(s, "\ufeff"))
	if s == "" {
		return "", errors.New("пустая строка")
	}
	// CRLF считается одним разделителем; пробелы между словами сохраняются.
	s = strings.ReplaceAll(s, "\r\n", "\n")
	s = strings.Map(func(ch rune) rune {
		if unicode.IsSpace(ch) {
			return ' '
		}
		return ch
	}, s)
	if IsMorse(s) {
		for _, code := range strings.Fields(s) {
			if morse.MorseToRune(code) == 0 {
				return "", fmt.Errorf("неизвестный код Морзе: %q", code)
			}
		}
		return morse.ToText(s), nil
	}
	for _, ch := range s {
		if ch != ' ' && morse.RuneToMorse(ch) == "" {
			return "", fmt.Errorf("неподдерживаемый символ: %q", ch)
		}
	}
	return morse.ToMorse(s), nil
}
