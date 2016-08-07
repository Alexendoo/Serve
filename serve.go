package main

import (
	"flag"
	"fmt"
	"html/template"
	"io"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

var (
	host       string
	port       string
	version    = "HEAD"
	linkTmpl   = template.Must(template.New("link").Parse("<a href=\"{{.}}\">{{.}}</a>\n"))
	headerTmpl = template.Must(template.New("header").Parse("<h2>{{.}}</h2>"))
)

const (
	usage = `
NAME:
   Serve - HTTP server for files spanning multiple directories https://git.io/serve

USAGE:
   %s [OPTION]... [DIR]...

VERSION:
   %s

OPTIONS:
   {{range .VisibleFlags}}{{.}}
   {{end}}
`
)

func main() {
	flags := getFlags()
	serve(flags)
}

func getFlags() *flag.FlagSet {
	flags := flag.NewFlagSet("flags", flag.ContinueOnError)
	flags.Usage = func() {
		usageName := filepath.Base(os.Args[0])
		fmt.Printf(usage, usageName, version)
	}
	flags.StringVar(&port, "port", "8080", "")
	flags.StringVar(&port, "p", "8080", "")
	flags.StringVar(&host, "host", "localhost", "")
	err := flags.Parse(os.Args[1:])
	if err == flag.ErrHelp {
		os.Exit(0)
	}
	if err != nil {
		os.Exit(1)
	}
	return flags
}

func serve(flags *flag.FlagSet) {
	dirs := make([]string, flags.NArg())
	for i := range dirs {
		dirs[i] = flags.Arg(i)
	}
	if len(dirs) == 0 {
		dirs = []string{"."}
	}
	http.HandleFunc("/", makeHandler(dirs))
	address := net.JoinHostPort(host, port)
	log.Fatal(http.ListenAndServe(address, nil))
}

func makeHandler(dirs []string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if tryFiles(w, r, dirs) {
			return
		}
		tryDirs(w, r, dirs)
	}
}

func tryFiles(w http.ResponseWriter, r *http.Request, dirs []string) bool {
	requestPath := r.URL.Path
	for _, dir := range dirs {
		filePath := filepath.Join(dir, requestPath)
		indexPath := filepath.Join(filePath, "index.html")
		if tryFile(w, r, filePath) || tryFile(w, r, indexPath) {
			return true
		}
	}
	return false
}

func tryFile(w http.ResponseWriter, r *http.Request, filePath string) bool {
	stat, statErr := os.Stat(filePath)
	if statErr != nil || stat.IsDir() {
		return false
	}

	file, fileErr := os.Open(filePath)
	if fileErr != nil {
		return false
	}

	http.ServeContent(w, r, stat.Name(), stat.ModTime(), file)
	return true
}

// TODO:
// - trailing slash on dir
// - file modes/sizes?
// - insert ..
// - append dimmed requestPath to dir

func tryDirs(w http.ResponseWriter, r *http.Request, dirs []string) bool {
	if !strings.Contains(r.Header.Get("Accept"), "text/html") {
		return false
	}

	requestPath := r.URL.Path
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	io.WriteString(w, `<!DOCTYPE html><html><head><meta charset="UTF-8"><title>indexes</title></head><body><pre>`)
	for _, dir := range dirs {
		dirPath := filepath.Join(dir, requestPath)
		contents, err := ioutil.ReadDir(dirPath)
		if err != nil {
			continue
		}
		headerTmpl.Execute(w, dir)
		for _, fileInfo := range contents {
			linkTmpl.Execute(w, fileInfo.Name())
		}
	}
	io.WriteString(w, `</pre></body></html>`)
	return false
}
