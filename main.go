package main

import (
	"errors"
	"io"
	"io/fs"
	"log"
	"net/http"
	"os"
	"path"
	"strings"

	flag "github.com/spf13/pflag"
)

var lport string

func init() {
	const (
		defaultPort = "8080"
		usage       = "The port to host the web server on."
	)
	flag.StringVar(&lport, "port", defaultPort, usage)
}

// TODO: 8443 https

// TODO: basic auth

// TODO: configure directories

func main() {
	http.Handle("/bin/", http.StripPrefix("/bin/", http.FileServer(http.Dir("bin"))))
	// http.HandleFunc("/bin", binHandler)
	http.HandleFunc("/chroot", chrootHandler)
	http.HandleFunc("/upload", uploadHandler)

	log.Printf("Serving at 0.0.0.0:%v", lport)
	log.Fatal(http.ListenAndServe(":"+lport, nil))
}

// func binHandler(w http.ResponseWriter, r *http.Request) {

// }

func chrootHandler(resp http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodPost { // Only accept POST requests
		resp.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	// Parse form
	// Up to 1Mb stored in memory
	if err := req.ParseMultipartForm(1024 * 1024); err != nil {
		log.Println("failed to parse multipart form:", err)
		resp.WriteHeader(http.StatusBadRequest)
		return
	} // AN: cba to DRY

	// Create chroot dir if it does not already exist. Only error on non-"file exists" error.
	if err := createFolderIfNotExists("chroot"); err != nil {
		log.Println("failed to create directory:", err)
		resp.WriteHeader(http.StatusInternalServerError)
		return
	}

	// Process files
	form := req.MultipartForm
	for filename, fileheaders := range form.File {
		newFile, err := createFileRecursive(path.Join("chroot", path.Clean(filename)))
		if err != nil {
			log.Println("failed to create file:", err)
			resp.WriteHeader(http.StatusInternalServerError)
			return
		}
		for _, header := range fileheaders {
			filepart, _ := header.Open()
			io.Copy(newFile, filepart)
		} // Note: it is important that the specification dictates that the headers are in order, and thus the parts of the file when reconstructed are not out of order
	}
	// TODO: // Process directories
	// for dirpath, dirnames := range form.Value {
	// 	for _, dir := range dirnames {
	// 		if strings.HasSuffix(dir, "/") {
	// 			if _, err := createFileRecursive(path.Join("chroot", dirpath, dir) + "/"); err != nil {
	// 				log.Println("failed to create file:", err)
	// 				resp.WriteHeader(http.StatusInternalServerError)
	// 				return
	// 			}
	// 		}
	// 	}
	// }

	resp.Write([]byte("Received file(s) successfully."))
}

func uploadHandler(resp http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodPost { // Only accept POST requests
		resp.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	// Parse form
	// Up to 1Mb stored in memory
	if err := req.ParseMultipartForm(1024 * 1024); err != nil {
		log.Println("failed to parse multipart form:", err)
		resp.WriteHeader(http.StatusBadRequest)
		return
	}

	// Create uploads dir if it does not already exist. Only error on non-"file exists" error.
	if err := createFolderIfNotExists("uploads"); err != nil {
		log.Println("failed to create directory:", err)
		resp.WriteHeader(http.StatusInternalServerError)
		return
	}

	// Process files
	form := req.MultipartForm
	for filename, fileheaders := range form.File {
		newFile, err := os.Create(path.Join(
			"uploads",
			strings.ReplaceAll(path.Clean(filename), "/", "_"),
		))
		if err != nil {
			log.Println("failed to create file:", err)
			resp.WriteHeader(http.StatusInternalServerError)
			return
		}
		for _, header := range fileheaders {
			filepart, _ := header.Open()
			io.Copy(newFile, filepart)
		} // Note: it is important that the specification dictates that the headers are in order, and thus the parts of the file when reconstructed are not out of order
	}

	resp.Write([]byte("Received file successfully."))
}

// TODO: cache / memoize

// createFileRecursive creates a file at the specified path,
// creating intermediary folders along the way. If the provided filepath
// ends in a slash, it will return nil file and nil error, creating the directories
// but not creating any file at the end.
func createFileRecursive(filepath string) (*os.File, error) {
	// Save current directory and come back to it
	olddir, err := os.Getwd()
	if err != nil {
		return nil, err
	}
	defer os.Chdir(olddir)

	dirs := strings.Split(strings.TrimPrefix(path.Clean(filepath), "/"), "/")
	for _, dir := range dirs[:len(dirs)-1] {
		if err := createFolderIfNotExists(dir); err != nil {
			return nil, err
		}
		if err := os.Chdir(dir); err != nil {
			return nil, err
		}
	}
	filename := dirs[len(dirs)-1]
	if filename != "" {
		return os.Create(filename)
	} else {
		return nil, nil
	}
}

// createFolderIfNotExists creates a directory if it does not already exist.
// Does not throw an error due to the directory already existing
func createFolderIfNotExists(folder string) error {
	err := os.Mkdir(folder, os.ModeDir|os.ModePerm)
	if errors.Is(err, fs.ErrExist) {
		return nil
	}
	return err
}
