package service

import (
	"errors"
	"strings"

	"github.com/Yandex-Practicum/go1fl-sprint6-final/pkg/morse"
)

func IsMorse(s string) bool {
	if len(strings.TrimSpace(s)) == 0 {
		return false
	}
	return strings.Trim(s, ".- \t\n\r") == ""
}

func Convert(s string) (string, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return "", errors.New("пустая строка")
	}
	if IsMorse(s) {
		return morse.ToText(s), nil
	}
	return morse.ToMorse(s), nil
}
