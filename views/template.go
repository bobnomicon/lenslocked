package views

import (
	"bytes"
	"errors"
	"fmt"
	"html/template"
	"io"
	"io/fs"
	"log"
	"net/http"
	"path/filepath"

	"github.com/bobnomicon/lenslocked/context"
	"github.com/bobnomicon/lenslocked/models"
	"github.com/gorilla/csrf"
)

type Template struct {
	htmlTpl *template.Template
}

type public interface {
	Public() string
}

func getErrMessages(errs ...error) []string {
	var msgs []string

	for _, err := range errs {
		var pubErr public

		if errors.As(err, &pubErr) {
			msgs = append(msgs, pubErr.Public())
		} else {
			fmt.Println(err)
			msgs = append(msgs, "Something went wrong.")
		}
	}

	return msgs
}

func (t Template) Execute(w http.ResponseWriter, r *http.Request, data any, errs ...error) {
	// Clone the template
	htmlTpl, err := t.htmlTpl.Clone()
	if err != nil {
		log.Printf("cloning template: %v", err)
		http.Error(w, "There was an error rendering the page.", http.StatusInternalServerError)
		return
	}

	// Get error messages if any
	errMsgs := getErrMessages(errs...)

	// Replace placholder funcs
	htmlTpl = htmlTpl.Funcs(
		template.FuncMap{
			"csrfField": func() template.HTML {
				return csrf.TemplateField(r)
			},
			"currentUser": func() *models.User {
				return context.User(r.Context())
			},
			"errors": func() []string {
				return errMsgs
			},
		},
	)
	
	// Set the content type header
	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	// Create a buffer to store the final result from template execution
	var buf bytes.Buffer 

	// Execute the cloned template	
	err = htmlTpl.Execute(&buf, data)
	if err != nil {
		log.Printf("executing template: %v", err)
		http.Error(w, "There was an error executing the template.", http.StatusInternalServerError)
		return
	}

	// Copy buffer to response writer if no errors (this is not efficient for large templates)
	io.Copy(w, &buf)
}

func ParseFS(fs fs.FS, pattern ...string) (Template, error) {
	htmlTpl := template.New(filepath.Base(pattern[0]))

	// Placeholder funcs
	htmlTpl = htmlTpl.Funcs(
		template.FuncMap{
			"csrfField": func() (template.HTML, error) {
				return "", fmt.Errorf("csrfField not implemented")
			},
			"currentUser": func() (*models.User, error) {
				return nil, fmt.Errorf("currentUser not implemented")
			},
			"errors": func() []string {
				return nil
			},
		},
	)

	// Parse the template
	htmlTpl, err := htmlTpl.ParseFS(fs, pattern...)
	if err != nil {
		return Template{}, fmt.Errorf("parsing template: %w", err)
	}

	return Template{
		htmlTpl: htmlTpl,
	}, nil
}

func Must(t Template, err error) Template {
	if err != nil {
		panic(err)
	}
	return t
}
