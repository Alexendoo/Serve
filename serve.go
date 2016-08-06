package main

import (
	"html/template"
	"io"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"os"
	"path"
	"strings"

	"github.com/urfave/cli"
)

func main() {
	cli.AppHelpTemplate = `NAME:
   Serve - HTTP server for files spanning multiple directories

USAGE:
   {{.HelpName}} [OPTION]... [DIR]...
   {{if .Version}}{{if not .HideVersion}}
VERSION:
   {{.Version}}
   {{end}}{{end}}{{if len .Authors}}
AUTHOR(S):
   {{range .Authors}}{{.}}{{end}}
   {{end}}{{if .VisibleFlags}}
OPTIONS:
   {{range .VisibleFlags}}{{.}}
   {{end}}{{end}}{{if .Copyright}}
COPYRIGHT:
   {{.Copyright}}
   {{end}}
`
	app := cli.NewApp()
	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:  "host",
			Usage: "bind to `address`",
			Value: "localhost",
		},
		cli.StringFlag{
			Name:  "port, p",
			Usage: "bind to `port`",
			Value: "8080",
		},
	}
	app.Action = action
	app.Run(os.Args)
}

func action(c *cli.Context) error {
	dirs := make([]string, c.NArg())
	for i := range dirs {
		dirs[i] = c.Args().Get(i)
	}
	if len(dirs) == 0 {
		dirs = []string{"."}
	}
	http.HandleFunc("/", makeHandler(dirs))
	address := net.JoinHostPort(c.String("host"), c.String("port"))
	log.Fatal(http.ListenAndServe(address, nil))
	return nil
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
		filePath := path.Join(dir, requestPath)
		indexPath := path.Join(filePath, "index.html")
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
	linkTmpl := template.Must(template.New("link").Parse("<a href=\"{{.}}\">{{.}}</a>\n"))
	headerTmpl := template.Must(template.New("header").Parse("<h2>{{.}}</h2>"))
	for _, dir := range dirs {
		dirPath := path.Join(dir, requestPath)
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
