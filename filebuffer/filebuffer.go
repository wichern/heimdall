package filebuffer

import (
	"fmt"
	"html/template"
	"io/ioutil"
	"os"
	"time"
)

type file struct {
	body         []byte
	template_ptr *template.Template
	path         string
	mtime        time.Time
}

type files struct {
	// Map path to file object.
	files map[string]file
}

func Get() files {
	return files{map[string]file{}}
}

func (files files) Get(path string) (file, error) {
	// Check whether we already loaded the file.
	f, ok := files.files[path]

	if !ok {
		return files.load(path)
	}

	// Check whether the file has been updated.
	fi, err := os.Stat(path)
	if err != nil {
		return f, err
	}
	diff := f.mtime.Sub(fi.ModTime())
	if diff < time.Duration(0) {
		fmt.Println("Reloading " + path)
		return files.load(path)
	}

	return f, nil
}

func (files files) load(path string) (file, error) {
	f := file{nil, nil, path, time.Time{}}

	fi, err := os.Stat(path)
	if err != nil {
		return f, err
	}
	f.mtime = fi.ModTime()

	files.files[path] = f
	return f, nil
}

func (file file) GetTemplate() *template.Template {
	if file.template_ptr == nil {
		file.template_ptr = template.Must(template.ParseFiles(file.path))
	}
	return file.template_ptr
}

func (file file) GetBody() []byte {
	if file.body == nil {
		var err error
		file.body, err = ioutil.ReadFile(file.path)
		if err != nil {
			fmt.Println(err)
			return nil
		}
	}
	return file.body
}
