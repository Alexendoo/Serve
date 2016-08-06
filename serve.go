package main

import (
	"log"
	"net"
	"net/http"
	"os"
	"path"
	"strings"

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
	cli.AppHelpTemplate = helpTemplate
	app := cli.NewApp()
	app.Name = "Serve"
	app.Usage = "HTTP server for files spanning multiple directories"
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
	file, fileErr := os.Open(filePath)

	opened := statErr == nil && fileErr == nil

	if opened && !stat.IsDir() {
		http.ServeContent(w, r, stat.Name(), stat.ModTime(), file)
		return true
	}
	return false
}

func tryDirs(w http.ResponseWriter, r *http.Request, dirs []string) bool {
	if !strings.Contains(r.Header.Get("Accept"), "text/html") {
		return false
	}

	requestPath := r.URL.Path
	for _, dir := range dirs {
		log.Printf("%s in %s", requestPath, dir)
	}
	return false
}
