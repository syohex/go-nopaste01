package main

import (
	"crypto/sha1"
	"fmt"
	"html/template"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
)

const indexHTML = `<!DOCTYPE html>
<html>
  <head>
    <meta charset="utf-8">
    <title>Nopaste</title>
    </head>
  <body>
  <h1>Nopaste in Go</h1>
    <form method="post">
    <p>
       <textarea id="snippet" name="snippet" rows="30" cols="100"></textarea>
    </p>
    <p>
       <input type="submit" value="Paste it">
    </p>
    </form>
  </body>
</html>
`

var indexTmpl = template.Must(template.New("indexTmpl").Parse(indexHTML))

func calcID(data []byte) string {
	hex := fmt.Sprintf("%x", sha1.Sum(data))
	return hex[0:8]
}

func savedFileName(id string) string {
	return filepath.Join("snippets", id+".txt")
}

func internalServerError(err error, w http.ResponseWriter) {
	log.Println(err)
	code := http.StatusInternalServerError
	http.Error(w, http.StatusText(code), code)
}

func saveSnippet(w http.ResponseWriter, req *http.Request) {
	data := []byte(req.FormValue("snippet"))
	if len(data) == 0 {
		http.Redirect(w, req, "/", http.StatusFound)
		return
	}

	id := calcID(data)
	name := savedFileName(id)
	if err := ioutil.WriteFile(name, data, 0644); err != nil {
		internalServerError(err, w)
		return
	}

	log.Printf("Save to '%s'", name)
	http.Redirect(w, req, "/"+id, http.StatusFound)
}

func indexHandler(w http.ResponseWriter, req *http.Request) {
	path := req.URL.Path

	if path == "/" {
		if req.Method == "POST" {
			saveSnippet(w, req)
			return
		} else if err := indexTmpl.Execute(w, nil); err != nil {
			internalServerError(err, w)
			return
		}
		return
	}

	path = path[1:]
	file := savedFileName(path)
	f, err := os.Open(file)
	if err != nil {
		log.Println(err)
		http.NotFound(w, req)
		return
	}
	defer f.Close()

	w.Header().Add("Content-Type", "text/plain")
	io.Copy(w, f)
}

func main() {
	if err := os.Mkdir("snippets", 01777); err != nil {
		log.Println(err)
	}

	http.HandleFunc("/", indexHandler)
	http.ListenAndServe(":5000", nil)
}
