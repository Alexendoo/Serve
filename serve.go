package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"os"
	"time"

	"net/http"

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
	paths := make([]string, c.NArg())
	for i := range paths {
		paths[i] = c.Args().Get(i)
	}
	http.HandleFunc("/", makeHandler(paths))
	log.Fatal(http.ListenAndServe("localhost:8080", nil))
	return nil
}

func makeHandler(paths []string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "Test: %q\n", r.URL.Path[1:])
		for _, path := range paths {
			dir, _ := ioutil.ReadDir(path)
			fmt.Fprintf(w, "\n%q\n", dir)
		}
	}
}

func ephemeralPort() int {
	return rand.Int()%16384 + 49152
}
