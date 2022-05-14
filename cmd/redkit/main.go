package main

import (
	"archive/zip"
	"bufio"
	"io"
	"io/fs"
	"log"
	"net/http"
	"os"
	"path"
	"strings"

	. "github.com/nobodyawesomer/results"
	flag "github.com/spf13/pflag"
)

var port string
var kitsDir string

func init() {
	flag.StringVarP(&port, "port", "p", "8080", "The port to host it on.")
	flag.StringVarP(&kitsDir, "kits", "k", "kits", "The directory of hosted kits.")
}

func main() {
	flag.Parse()

	http.HandleFunc("/kit/", serveFiles)

	log.Println("Serving from 0.0.0.0:" + port)
	log.Fatal(http.ListenAndServe("0.0.0.0:"+port, nil))
}

func serveFiles(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet { // Only handle GETs
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	// Split path
	log.Println("urlpath: ", r.URL.Path)
	requestPath := strings.Split(r.URL.Path, "/")
	if len(requestPath) < 1 {
		http.NotFound(w, r)
		return
	}
	// "/kit/linux" -> ["", "kit", "linux"]

	if requestPath[1] != "kit" {
		http.NotFound(w, r)
		return
	} // guarantee GET /kit*

	// Get kits directory
	log.Println(Unwrap(os.Getwd()), kitsDir)
	subDirs := Try(os.ReadDir(kitsDir)).
		Catch(os.ErrPermission, func(r *[]fs.DirEntry, err error) {
			log.Println("No permission to access", (*r)[len(*r)-1].Name(), "directory")
		}).
		Catch(os.ErrNotExist, func(r *[]fs.DirEntry, err error) {
			log.Fatal("Could not access kits directory: ", err)
		}).
		Unwrap()

	// Handle GET /kit case - if ./kits has a ./kits/bin, then serve it zipped.
	// TODO impl

	// Handle any /kit/* case - if ./kits/linux has a ./kits/linux/bin, then serve it zipped.
	if len(requestPath) < 3 {
		http.NotFound(w, r)
		return
	} // guarantee GET /kit/something
	requestedKit := requestPath[2]

	for _, kitDir := range subDirs {
		if kitDir.Name() != requestedKit {
			continue
		}
		log.Println("Seeing ", kitDir)
		binPath := path.Join(kitsDir, kitDir.Name(), "bin")
		_, err := os.Stat(binPath)
		if err != nil {
			log.Printf("Failed to stat bin directory: %v", err)
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(err.Error()))
			return
		}
		log.Printf("Handling path %v from request to %v", binPath, requestPath)
		// err.... does this bufferthing work? i only want to send back one response
		zipDirFull(binPath, bufio.NewWriter(w))
		// using w automatically writes
		return
	}

	// All else:
	http.NotFound(w, r)
}

// zipDirFull zips a directory, pulling in the real files symlinks stand for.
func zipDirFull(basePath string, buf io.Writer) {
	log.Println("Making zip.Writer from path", basePath)
	archive := zip.NewWriter(buf)

	Check(fs.WalkDir(os.DirFS(basePath), ".", func(filePath string, d fs.DirEntry, err error) error {
		if d == nil {
			return err // fs.Stat on root directory failed.
		}
		fil := followIfSymlink(Unwrap(os.Open(path.Join(basePath, filePath))))
		filStat := Unwrap(fil.Stat())
		if filStat.Size() < 1 || // Don't zip empty files
			filStat.IsDir() { // Don't zip dirs
			return nil
		}
		zipf := Unwrap(archive.Create(filePath))
		log.Printf("Filling archive with %v", filePath)
		io.Copy(zipf, fil)
		return nil
	}))

	Check(archive.Close())
}

// followIfSymlink takes a file and returns the source if it is a symlink.
// Otherwise, it just returns the input.
func followIfSymlink(maybeSym *os.File) *os.File {
	if Unwrap(maybeSym.Stat()).Mode()&os.ModeSymlink != 1 {
		return maybeSym
	}
	// It is a symlink.
	defer os.Chdir(Unwrap(os.Getwd()))
	Check(maybeSym.Chdir())
	realFileName := Unwrap(os.Readlink(maybeSym.Name()))

	// Recursively follow symlinks
	return followIfSymlink(Unwrap(os.Open(realFileName)))
}
