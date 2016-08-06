package main

import (
	"errors"
	"log"
	"math/rand"
	"net/http"
	"os"
	"path"
	"strings"
	"time"

	"github.com/urfave/cli"
)

const helpTemplate = `NAME:
   {{.Name}} - {{.Usage}}
USAGE:
   {{if .UsageText}}{{.UsageText}}{{else}}{{.HelpName}} {{if .VisibleFlags}}[global options]{{end}}{{if .Commands}} command [command options]{{end}} {{if .ArgsUsage}}{{.ArgsUsage}}{{else}}[arguments...]{{end}}{{end}}
   {{if .Version}}{{if not .HideVersion}}
VERSION:
   {{.Version}}
   {{end}}{{end}}{{if len .Authors}}
AUTHOR(S):
   {{range .Authors}}{{.}}{{end}}
   {{end}}{{if .VisibleFlags}}
GLOBAL OPTIONS:
   {{range .VisibleFlags}}{{.}}
   {{end}}{{end}}{{if .Copyright}}
COPYRIGHT:
   {{.Copyright}}
   {{end}}
`

func main() {
	rand.Seed(time.Now().UnixNano())
	cli.AppHelpTemplate = helpTemplate
	app := cli.NewApp()
	app.Name = "Serve"
	app.Usage = "HTTP server for files spanning multiple directories"
	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:  "port, p",
			Usage: "`port` to bind server to (default: random)",
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
	http.HandleFunc("/", makeHandler(dirs))
	log.Fatal(http.ListenAndServe("localhost:8080", nil))
	return nil
}

func makeHandler(dirs []string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		err := tryFiles(w, r, dirs)
		if err == nil {
			return
		}
		log.Println("hello")
		tryDirs(w, r, dirs)
	}
}

func tryFiles(w http.ResponseWriter, r *http.Request, dirs []string) error {
	requestPath := r.URL.Path
	for _, dir := range dirs {
		tryFile := path.Join(dir, requestPath)
		stat, statErr := os.Stat(tryFile)
		file, fileErr := os.Open(tryFile)

		opened := statErr == nil && fileErr == nil

		if opened && !stat.IsDir() {
			http.ServeContent(w, r, stat.Name(), stat.ModTime(), file)
			return nil
		}
	}
	return errors.New("no files found")
}

func tryDirs(w http.ResponseWriter, r *http.Request, dirs []string) {
	requestPath := r.URL.Path
	htmlReqeust := strings.Contains(r.Header.Get("Accept"), "text/html")
	log.Println(requestPath, htmlReqeust)
	for _, dir := range dirs {
		log.Printf("%s in %s", requestPath, dir)
	}
}

func ephemeralPort() int {
	return rand.Int()%16384 + 49152
}
