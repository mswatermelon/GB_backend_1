package main_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	. "github.com/mswatermelon/GB_backend_1/file-server"
	"github.com/mswatermelon/GB_backend_1/file-server/upload"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

func TestGetHandler(t *testing.T) {
	// Создаем запрос с указанием нашего хендлера. Так как мы тестируем GET-эндпоинт
	// то нам не нужно передавать тело, поэтому третьим аргументом передаем nil
	req, err := http.NewRequest(http.MethodGet, "/?name=John", nil)
	if err != nil {
		t.Fatal(err)
	}
	// Мы создаем ResponseRecorder(реализует интерфейс http.ResponseWriter)
	// и используем его для получения ответа
	rr := httptest.NewRecorder()
	handler := &Handler{}
	// Наш хендлер соответствует интерфейсу http.Handler, а значит
	// мы можем использовать ServeHTTP и напрямую указать
	// Request и ResponseRecorder
	handler.ServeHTTP(rr, req)
	// Проверяем статус-код ответа
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusOK)
	}
	// Проверяем тело ответа
	expected := `Parsed query-param with key "name": John`
	if rr.Body.String() != expected {
		t.Errorf("handler returned unexpected body: got %v want %v",
			rr.Body.String(), expected)
	}
}

func TestUploadHandler(t *testing.T) {
	// открываем файл, который хотим отправить
	file, _ := os.Open("testfile")
	defer file.Close()
	// действия, необходимые для того, чтобы засунуть файл в запрос
	// в качестве мультипарт-формы
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, _ := writer.CreateFormFile("file", filepath.Base(file.Name()))
	io.Copy(part, file)
	writer.Close()
	// опять создаем запрос, теперь уже на /upload эндпоинт
	req, _ := http.NewRequest(http.MethodPost, "/upload", body)
	req.Header.Add("Content-Type", writer.FormDataContentType())
	// создаем ResponseRecorder
	rr := httptest.NewRecorder()
	// создаем заглушку файлового сервера. Для прохождения тестов
	// нам достаточно чтобы он возвращал 200 статус
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "ok!")
	}))
	defer ts.Close()
	uploadHandler := &upload.Handler{
		UploadDir: "upload",
		// таким образом мы подменим адрес файлового сервера
		// и вместо реального, хэндлер будет стучаться на заглушку
		// которая всегда будет возвращать 200 статус, что нам и нужна
		HostAddr: ts.URL,
	}
	// опять же, вызываем ServeHTTP у тестируемого обработчика
	uploadHandler.ServeHTTP(rr, req)
	// Проверяем статус-код ответа
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusOK)
	}
	expected := `testfile`
	if !strings.Contains(rr.Body.String(), expected) {
		t.Errorf("handler returned unexpected body: got %v want %v",
			rr.Body.String(), expected)
	}
}

func test(m interface{}) bool {
	rt := reflect.TypeOf(m)

	switch rt.Kind() {
	case reflect.Slice:
		return true
	case reflect.Array:
		return true
	default:
		return false
	}
}

func TestFileHandler(t *testing.T) {
	req, _ := http.NewRequest(http.MethodGet, "/upload", nil)
	rr := httptest.NewRecorder()
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "ok!")
	}))
	defer ts.Close()
	fileHandler := &upload.FileServeHandler{
		Dir: "upload",
	}
	fileHandler.ServeHTTP(rr, req)
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusOK)
	}
	var target []upload.FileInfo
	err := json.NewDecoder(rr.Body).Decode(&target)
	if err != nil {
		t.Errorf("Unable to parse the answer")
	}
	if !test(target) {
		t.Errorf("handler returned unexpected body: got %v want %v",
			rr.Body.String(), target)
	}
}

func TestFileHandlerWithFilter(t *testing.T) {
	req, _ := http.NewRequest(http.MethodGet, "/upload?ext=.json", nil)
	rr := httptest.NewRecorder()
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "ok!")
	}))
	defer ts.Close()
	fileHandler := &upload.FileServeHandler{
		Dir: "upload",
	}
	fileHandler.ServeHTTP(rr, req)
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusOK)
	}
	m := rr.Body.String()
	println(m)
	var target []upload.FileInfo
	err := json.NewDecoder(rr.Body).Decode(&target)
	switch {
	case err == io.EOF:
		// empty body
	case err != nil:
		t.Errorf(fmt.Errorf("Unable to parse the answer: %w", err).Error())
		return
	}
	if !test(target) {
		t.Errorf("handler returned unexpected body: got %v want %v",
			rr.Body.String(), target)
	}
}
