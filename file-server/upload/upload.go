package upload

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"path/filepath"
	"time"
)

type Handler struct {
	HostAddr  string
	UploadDir string
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	file, header, err := r.FormFile("file")
	if err != nil {
		http.Error(w, "Unable to read file", http.StatusBadRequest)
		return
	}
	defer file.Close()
	data, err := ioutil.ReadAll(file)
	if err != nil {
		http.Error(w, "Unable to read file", http.StatusBadRequest)
		return
	}
	fileName := header.Filename
	filePath := h.UploadDir + "/" + fileName
	err = ioutil.WriteFile(filePath, data, 0777)
	if err != nil {
		log.Println(err)
		http.Error(w, "Unable to save file", http.StatusInternalServerError)
		return
	}
	fmt.Fprintf(w, "File %s has been successfully uploaded", fileName)

	fileLink := h.HostAddr + "/" + header.Filename
	req, err := http.NewRequest(http.MethodHead, fileLink, nil)
	if err != nil {
		log.Println(err)
		http.Error(w, "Unable to check file", http.StatusInternalServerError)
		return
	}
	cli := &http.Client{}
	resp, err := cli.Do(req)
	if err != nil {
		log.Println(err)
		http.Error(w, "Unable to check file", http.StatusInternalServerError)
		return
	}
	if resp.StatusCode != http.StatusOK {
		log.Println(err)
		http.Error(w, "Unable to save file", http.StatusInternalServerError)
		return
	}
	fmt.Fprintln(w, fileLink)
}

type FileInfo struct {
	Name      string `json:"name"`
	Extension string `json:"extension"`
	Size      int64  `json:"size"`
}

type FileServeHandler struct {
	dir string
}

func (h *FileServeHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	files, err := ioutil.ReadDir(h.dir)
	if err != nil {
		log.Println(err)
		http.Error(w, "Can not open file's directory", http.StatusInternalServerError)
		return
	}

	fileInfo := make([]FileInfo, len(files))
	for _, file := range files {
		fileInfo = append(fileInfo, FileInfo{
			Name:      file.Name(),
			Extension: filepath.Ext(h.dir + "/" + file.Name()),
			Size:      file.Size(),
		})
	}
	jsonResp, err := json.Marshal(fileInfo)
	if err != nil {
		log.Println(err)
		http.Error(w, "Can not create a list of files", http.StatusInternalServerError)
		return
	}
	_, err = w.Write(jsonResp)
	if err != nil {
		log.Println(err)
		http.Error(w, "Unable to write the file's list", http.StatusInternalServerError)
	}
}

func main() {
	uploadHandler := &Handler{
		UploadDir: "upload",
	}
	fileServeHandler := &FileServeHandler{
		dir: "upload",
	}
	http.Handle("/upload", uploadHandler)
	http.Handle("/files", fileServeHandler)

	dirToServe := http.Dir(uploadHandler.UploadDir)
	fs := &http.Server{
		Addr:         ":8080",
		Handler:      http.FileServer(dirToServe),
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}
	fs.ListenAndServe()
}
