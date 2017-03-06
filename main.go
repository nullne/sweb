package main

import (
	"bytes"
	"flag"
	"io"
	"log"
	"net/http"
	"os"
	"path"
	"strings"
	"sync"
	"time"

	"github.com/toqueteos/webbrowser"
)

var (
	doc  *document
	lock sync.RWMutex

	documentPath = flag.String("f", "api-spec.yaml", "the full path to the document being edited")
	staticPath   = flag.String("s", "", "all files under the path will be served")
	backendPort  = flag.String("p", "8765", "port for editor's http backend")
	editorPath   = flag.String("se", "builtin", "the full path to swagger-editor installation")
)

type document struct {
	sync.RWMutex
	path  string
	saved bool
	buf   *bytes.Buffer
}

func init() {
	flag.Parse()
	if *staticPath == "" {
		*staticPath = path.Dir(*documentPath)
	}
	doc = &document{
		buf:  &bytes.Buffer{},
		path: *documentPath,
	}

	go doc.doSync()
}

func (doc *document) doSync() {
	err := doc.open()
	if err != nil {
		log.Println(err)
		return
	}

	tick := time.NewTicker(2 * time.Second).C
	for {
		select {
		case <-tick:
			if !doc.saved {
				doc.save()
			}
		}
	}
}

func (doc *document) open() error {
	doc.Lock()
	f, err := os.Open(doc.path)
	if err != nil {
		if os.IsNotExist(err) {
			f, err = os.Create(doc.path)
		}

		if err != nil {
			return err
		}

	}

	defer f.Close()
	defer doc.Unlock()

	io.Copy(doc.buf, f)
	return nil
}

func (doc *document) save() error {
	doc.RLock()
	f, err := os.Create(doc.path)
	if err != nil {
		return err
	}

	defer f.Close()

	n, err := f.Write(doc.buf.Bytes())
	if err != nil {
		return err
	}

	doc.saved = true
	doc.RUnlock()

	log.Printf("%v bytes saved\n", n)
	return nil
}

func handleBackend(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		doc.RLock()
		_, err := w.Write(doc.buf.Bytes())
		if err != nil {
			log.Println(err)
		}
		doc.RUnlock()

	case http.MethodPut:
		doc.Lock()
		doc.buf.Reset()
		_, err := io.Copy(doc.buf, r.Body)
		if err != nil {
			log.Println(err)
		}
		doc.saved = false
		doc.Unlock()
	}
}

func handleApp(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path
	if path == "/" {
		path = "/index.html"
	}

	data, err := Asset("swagger-editor" + path)
	if err != nil {
		log.Println(path)
		http.Error(w, "resource not found"+path, http.StatusNotFound)
	}

	contentType := http.DetectContentType(data)
	w.Header().Set("Content-Type", contentType)

	if strings.HasSuffix(path, ".js") {
		w.Header().Set("Content-Type", "application/javascript")
	}

	if strings.HasSuffix(path, ".css") {
		w.Header().Set("Content-Type", "text/css")
	}

	w.Write(data)
}

func fullCORS(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Access-Control-Allow-Origin", "*")
		w.Header().Add("Access-Control-Allow-Methods", "GET, POST, DELETE, PUT, PATCH, OPTIONS")
		w.Header().Add("iAccess-Control-Allow-Headers", "Content-Type, api_key, Authorization")
		h.ServeHTTP(w, r)
	})
}

func main() {
	webbrowser.Open("http://localhost:" + *backendPort)

	http.HandleFunc("/backend", handleBackend)
	http.Handle("/static/", http.StripPrefix("/static/", fullCORS(http.FileServer(http.Dir(*staticPath)))))
	if *editorPath == "builtin" {
		http.HandleFunc("/", handleApp)
	} else {
		http.Handle("/", http.FileServer(http.Dir(*editorPath)))
	}

	log.Fatal(http.ListenAndServe(":8765", nil))
}
