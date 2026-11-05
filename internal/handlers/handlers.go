package handlers

import (
	"io"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/Yandex-Practicum/go1fl-sprint6-final/internal/service"
)

func Index(w http.ResponseWriter, r *http.Request) {
	content, err := os.ReadFile("index.html")
	if err != nil {
		http.Error(w, "Не удалось прочитать HTML-файл", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Write(content)
}

func Upload(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.Header().Set("Allow", http.MethodPost)
		http.Error(w, "Метод не поддерживается", http.StatusMethodNotAllowed)
		return
	}

	err := r.ParseMultipartForm(10 << 20)
	if err != nil {
		http.Error(w, "Ошибка при разборе формы", http.StatusInternalServerError)
		return
	}

	file, header, err := r.FormFile("myFile")
	if err != nil {
		http.Error(w, "Ошибка при получении файла", http.StatusInternalServerError)
		return
	}
	defer file.Close()
	ext := filepath.Ext(header.Filename)

	content, err := io.ReadAll(file)
	if err != nil {
		http.Error(w, "Ошибка при чтении файла", http.StatusInternalServerError)
		return
	}

	result, err := service.Convert(string(content))
	if err != nil {
		http.Error(w, "Ошибка при конвертации", http.StatusInternalServerError)
		return
	}
	filename := time.Now().UTC().String() + ext
	output, err := os.Create(filename)
	if err != nil {
		http.Error(w, "Ошибка при создании файла", http.StatusInternalServerError)
		return
	}
	defer output.Close()

	_, err = output.Write([]byte(result))
	if err != nil {
		http.Error(w, "Ошибка при записи в файл", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.Write([]byte(result))
}
