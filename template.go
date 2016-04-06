package render

import (
	"bytes"
	"fmt"
	"html/template"
	"net/http"
	"os"
	"path/filepath"
)

type Template struct {
	render *Render
	layout string
}

func (tmpl *Template) Execute(name string, context interface{}, request *http.Request, writer http.ResponseWriter) (err error) {
	if filename, ok := tmpl.findTemplate(name); ok {
		// filenames
		var filenames = []string{filename}
		var layout string
		if layout, ok = tmpl.findTemplate(filepath.Join("layouts", tmpl.layout)); ok {
			filenames = append(filenames, layout)
		}

		var result = map[string]interface{}{
			"Template": filename,
			"Result":   context,
		}

		// funcMaps
		var funcMap = tmpl.render.funcMaps
		funcMap["render"] = func(name string) (template.HTML, error) {
			var err error

			if filename, ok := tmpl.findTemplate(name); ok {
				var partialTemplate *template.Template
				result := bytes.NewBufferString("")
				if partialTemplate, err = template.New(filepath.Base(filename)).Funcs(funcMap).ParseFiles(filename); err == nil {
					partialTemplate.Execute(result, result)
					return template.HTML(result.String()), nil
				}
			} else {
				err = fmt.Errorf("failed to find template: %v", name)
			}

			return "", err
		}

		// parse templates
		var t *template.Template
		if t, err = template.New(filepath.Base(layout)).Funcs(tmpl.render.funcMaps).ParseFiles(filenames...); err == nil {
			err = t.Execute(writer, result)
		}
	}

	if err != nil {
		fmt.Printf("Got error when render template %v: %v\n", name, err)
	}
	return err
}

func (tmpl *Template) findTemplate(name string) (string, bool) {
	name = name + ".tmpl"
	for _, viewPath := range tmpl.render.ViewPaths {
		filename := filepath.Join(viewPath, name)
		if _, err := os.Stat(filename); !os.IsNotExist(err) {
			return filename, true
		}
	}
	return "", false
}
