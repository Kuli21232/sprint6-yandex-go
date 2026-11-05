package handlers

import (
	"net/http"
	"os"
)

func Index(w http.ResponseWriter, r *http.Request) {
	content, err := os.ReadFile("index.html")
	if err != nil {
		http.Error(w, "Не удалось прочитать HTML-файл", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Write(content)
}
