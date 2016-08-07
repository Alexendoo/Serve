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
	index      string
	noList     bool
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
   -p, --port     --  bind to port (default: 8080)
       --host     --  bind to host (default: localhost)
   -i, --index    --  serve all paths to index if file not found
       --no-list  --  disable file listings
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
	flags.StringVar(&index, "i", "", "")
	flags.StringVar(&index, "index", "", "")
	flags.BoolVar(&noList, "no-list", false, "")
	err := flags.Parse(os.Args[1:])
	if err == flag.ErrHelp {
		os.Exit(0)
	}
	if err != nil {
		os.Exit(1)
	}
	log.Printf("%q - %q - %q", port, host, index)
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
		server := fmt.Sprintf("serve/%s", version)
		w.Header().Set("Server", server)
		if !validRequest(r) {
			http.Error(w, "invalid path", http.StatusBadRequest)
			return
		}
		if tryFiles(w, r, dirs) {
			return
		}
		if !strings.Contains(r.Header.Get("Accept"), "text/html") {
			return
		}
		if len(index) > 0 && staticIndex(w, r) {
			return
		}
		if !noList && tryDirs(w, r, dirs) {
			return
		}
		http.NotFound(w, r)
	}
}

func validRequest(r *http.Request) bool {
	if !strings.Contains(r.URL.Path, "..") {
		return true
	}
	for _, field := range strings.FieldsFunc(r.URL.Path, isSlashRune) {
		if field == ".." {
			return false
		}
	}
	return true
}

func isSlashRune(r rune) bool { return r == '/' || r == '\\' }

func tryFiles(w http.ResponseWriter, r *http.Request, dirs []string) bool {
	for _, dir := range dirs {
		filePath := filepath.Join(dir, r.URL.Path)
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

func staticIndex(w http.ResponseWriter, r *http.Request) bool {
	file, fileErr := os.Open(index)
	stat, statErr := os.Stat(index)
	if fileErr != nil || statErr != nil {
		log.Println(fileErr)
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
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	io.WriteString(w, `<!DOCTYPE html><html><head><meta charset="UTF-8"><title>indexes</title></head><body><pre>`)
	for _, dir := range dirs {
		dirPath := filepath.Join(dir, r.URL.Path)
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
	return true
}
