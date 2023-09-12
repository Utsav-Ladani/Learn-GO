package main

import (
	"html/template"
	"log"
	"net/http"
	"os"
	"regexp"
	"strings"
)

type Page struct {
	Title string
	Body  []byte
}

func (p *Page) save() error {
	filename := p.Title + ".txt"
	return os.WriteFile("data/"+filename, p.Body, 0600)
}

func loadPage(title string) (*Page, error) {
	filename := title + ".txt"
	body, err := os.ReadFile("data/" + filename)
	if err != nil {
		return nil, err
	}
	return &Page{Title: title, Body: body}, nil
}

var templates = template.Must(template.ParseFiles("tmpl/item.html", "tmpl/all.html", "tmpl/edit.html", "tmpl/view.html"))
var validPath = regexp.MustCompile("^/(edit|save|view)/([a-zA-Z0-9]+)$")
var assetPath = regexp.MustCompile("^/asset/([a-zA-Z0-9]+.[a-zA-Z0-9]+)$")

func viewHandler(w http.ResponseWriter, r *http.Request, title string) {
	p, err := loadPage(title)
	if err != nil {
		http.Redirect(w, r, "/edit/"+title, http.StatusFound)
		return
	}
	renderTemplate(w, "view", p)
}

func viewAllHandler(w http.ResponseWriter, r *http.Request) {
	if "/" != r.URL.Path {
		http.NotFound(w, r)
		return
	}

	var Pages []*Page

	files, err := os.ReadDir("data")
	if err != nil {
		log.Fatal(err)
	}

	for _, file := range files {
		if file.IsDir() {
			continue
		}

		p, err := loadPage(strings.Split(file.Name(), ".")[0])
		if err != nil {
			log.Fatal(err)
		}

		Pages = append(Pages, p)
	}

	err = templates.ExecuteTemplate(w, "all.html", Pages)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func editHandler(w http.ResponseWriter, r *http.Request, title string) {
	p, err := loadPage(title)
	if err != nil {
		p = &Page{Title: title}
	}
	renderTemplate(w, "edit", p)
}

func saveHandler(w http.ResponseWriter, r *http.Request, title string) {
	body := r.FormValue("body")
	p := &Page{Title: title, Body: []byte(body)}
	err := p.save()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, "/view/"+title, http.StatusFound)
}

func renderTemplate(w http.ResponseWriter, tmpl string, p *Page) {
	err := templates.ExecuteTemplate(w, tmpl+".html", p)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func makeHandler(fn func(http.ResponseWriter, *http.Request, string)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		m := validPath.FindStringSubmatch(r.URL.Path)
		if m == nil {
			http.NotFound(w, r)
			return
		}
		fn(w, r, m[2])
	}
}

func assetHandler(w http.ResponseWriter, r *http.Request) {
	m := assetPath.FindStringSubmatch(r.URL.Path)

	if m == nil {
		http.NotFound(w, r)
		return
	}

	filename := m[1]

	asset, err := os.ReadFile("asset/" + filename)
	if err != nil {
		http.NotFound(w, r)
		return
	}

	contentType := map[string]string{
		"css":  "text/css",
		"js":   "text/javascript",
		"png":  "image/png",
		"jpg":  "image/jpeg",
		"jpeg": "image/jpeg",
		"gif":  "image/gif",
		"svg":  "image/svg+xml",
		"ico":  "image/x-icon",
		"mp4":  "video/mp4",
		"webm": "video/webm",
		"ogg":  "video/ogg",
		"mp3":  "audio/mpeg",
		"wav":  "audio/wav",
		"pdf":  "application/pdf",
		"zip":  "application/zip",
		"tar":  "application/x-tar",
		"gz":   "application/gzip",
	}

	extension := strings.Split(filename, ".")[1]

	if contentType[extension] != "" {
		w.Header().Set("Content-Type", contentType[extension])
	} else {
		http.NotFound(w, r)
	}

	w.Write(asset)
}

func main() {
	http.HandleFunc("/", viewAllHandler)
	http.HandleFunc("/asset/", assetHandler)

	http.HandleFunc("/view/", makeHandler(viewHandler))
	http.HandleFunc("/edit/", makeHandler(editHandler))
	http.HandleFunc("/save/", makeHandler(saveHandler))

	log.Fatal(http.ListenAndServe(":8080", nil))
}
