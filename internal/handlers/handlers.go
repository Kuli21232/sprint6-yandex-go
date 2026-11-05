package handlers

import (
	"errors"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/Yandex-Practicum/go1fl-sprint6-final/internal/service"
)

// MaxUploadSize задаёт максимальный размер загружаемого файла в байтах.
const MaxUploadSize = 10 << 20

const (
	multipartMemory = 1 << 20
	// Дополнительное место для заголовков multipart и остальных полей формы.
	maxRequestSize = MaxUploadSize + multipartMemory
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

	if r.ContentLength > maxRequestSize {
		http.Error(w, "Слишком большой запрос", http.StatusInternalServerError)
		return
	}
	r.Body = http.MaxBytesReader(w, r.Body, maxRequestSize)
	err := r.ParseMultipartForm(multipartMemory)
	if err != nil {
		var sizeErr *http.MaxBytesError
		if errors.As(err, &sizeErr) || errors.Is(err, multipart.ErrMessageTooLarge) {
			http.Error(w, "Слишком большой запрос", http.StatusInternalServerError)
			return
		}
		http.Error(w, "Ошибка при разборе формы", http.StatusInternalServerError)
		return
	}
	defer r.MultipartForm.RemoveAll()

	file, header, err := r.FormFile("myFile")
	if err != nil {
		http.Error(w, "Ошибка при получении файла", http.StatusInternalServerError)
		return
	}
	defer file.Close()
	if header.Size > MaxUploadSize {
		http.Error(w, "Файл превышает лимит 10 МиБ", http.StatusInternalServerError)
		return
	}
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
