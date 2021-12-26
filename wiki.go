package main

import (
	"os"
	"log"
	"net/http"
	"html/template"
	"io/ioutil"
	"regexp"
	"errors"
)

var templates = template.Must(template.ParseFiles(
	"./tmpl/edit.html", 
	"./tmpl/view.html",
	"./tmpl/index.html",
))
var validPath = regexp.MustCompile(
	"(?:^/(edit|save|view)/([a-zA-Z0-9]+)$|^/init$)",
)

type Page struct {
	Title string
	Body []byte
}

func (p *Page) save() error {
	filename := "./data/" + p.Title + ".txt"
	return os.WriteFile(filename, p.Body, 0600)
}

func loadPage(title string) (*Page, error) {
	filename := "./data/" + title + ".txt"
	body, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	return &Page{ Title: title, Body: body }, nil
}

func renderTemplate(w http.ResponseWriter, tmpl string, p *Page) {
	err := templates.ExecuteTemplate(w, tmpl+".html", p)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func getTitle(w http.ResponseWriter, r *http.Request) (string, error) {
	m := validPath.FindStringSubmatch(r.URL.Path)
	if m == nil {
		http.NotFound(w, r)
		return "", errors.New("invalid Page Title")
	}
	return m[2], nil // Title is the second subexpresison
}

func makeHandler(fn func(http.ResponseWriter, *http.Request, string)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		m := validPath.FindStringSubmatch(r.URL.Path)
		if m == nil {
			http.Redirect(w, r, "/", http.StatusFound)
			return
		}
		fn(w, r, m[2])
	}
}

func viewHandler(w http.ResponseWriter, r *http.Request, title string) {
	p, err := loadPage(title)
	if err != nil {
		http.Redirect(w, r, "/edit/"+title, http.StatusFound)
		return
	}
	renderTemplate(w, "view", p)
}

func editHandler(w http.ResponseWriter, r *http.Request, title string) {
	p, err := loadPage(title)
	if err != nil {
		p = &Page{ Title: title }
	}
	renderTemplate(w, "edit", p)
}

func saveHandler(w http.ResponseWriter, r *http.Request, title string) {
	body := r.FormValue("body")
	p := &Page{ Title: title, Body: []byte(body) }
	err := p.save()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, "/view/"+title, http.StatusFound)
}

func mainPageHandler(w http.ResponseWriter, r *http.Request) {
	files, err := ioutil.ReadDir("./data/")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	var fileNames []string
	for _, v := range files {
		fName := v.Name()
		fileNames = append(fileNames, fName[:len(fName) - len(".txt")])
	}
	templates.ExecuteTemplate(w, "index.html", fileNames)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func main() {
	http.HandleFunc("/view/", makeHandler(viewHandler))
	http.HandleFunc("/edit/", makeHandler(editHandler))
	http.HandleFunc("/save/", makeHandler(saveHandler))
	http.HandleFunc("/", mainPageHandler)
	http.HandleFunc("/index", mainPageHandler)

	log.Fatal(http.ListenAndServe(":8080", nil))
}
