package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"os"
	"path"
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
		requestPath := r.URL.Path
		if requestPath[len(requestPath)-1:] == "/" {
			requestPath += "index.html"
		}
		for _, dir := range dirs {
			dirDetails, _ := ioutil.ReadDir(dir)
			file, err := ioutil.ReadFile(path.Join(dir, requestPath))

			if err != nil {
				log.Println("?")
				w.Write(file)
				return
			}

			fmt.Fprintf(w, "\n%q\n", path.Join(dir, requestPath))
			fmt.Fprintf(w, "\n%q\n", http.Dir(requestPath))
			fmt.Fprintf(w, "\n%q\n", file)
			fmt.Fprintf(w, "\n%q\n", err)
			fmt.Fprintf(w, "\n%q\n", dirDetails)
		}
	}
}

func ephemeralPort() int {
	return rand.Int()%16384 + 49152
}
