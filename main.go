package main

import (
	"net/http"
	"os"

	"github.com/Sirupsen/logrus"
	"github.com/citadel/citadel/cluster"
	"github.com/citadel/citadel/scheduler"
	"github.com/codegangsta/cli"
	"github.com/gorilla/mux"
)

var (
	manager *cluster.Cluster
	logger  = logrus.New()
)

func createAPIHandler() http.Handler {
	r := mux.NewRouter()

	r.HandleFunc("/engines", engines).Methods("GET")
	r.HandleFunc("/engines", enginesAdd).Methods("POST")
	r.HandleFunc("/engines", enginesRemove).Methods("DELETE")

	r.HandleFunc("/containers", containers).Methods("GET")

	return r
}

func createHandler(dir string) http.Handler {
	var (
		mux         = http.NewServeMux()
		fileHandler = http.FileServer(http.Dir(dir))
	)

	mux.Handle("/api/", http.StripPrefix("/api", createAPIHandler()))
	mux.Handle("/", fileHandler)

	return mux
}

func main() {
	app := cli.NewApp()
	app.Name = "dockerui"
	app.Email = "crosbymichael@gmail.com"
	app.Author = "@crosbymichael"
	app.Version = "2"

	app.Flags = []cli.Flag{
		cli.StringFlag{Name: "addr", Value: "127.0.0.1:9000", Usage: "address to serve the API/UI"},
		cli.StringFlag{Name: "assets,a", Value: "assets/", Usage: "path to the assets directory"},
		cli.BoolFlag{Name: "debug", Usage: "enable debug output in logs"},
	}

	app.Before = func(context *cli.Context) error {
		if context.GlobalBool("debug") {
			logger.Level = logrus.DebugLevel
		}

		return nil
	}

	app.Action = func(context *cli.Context) {
		var (
			err     error
			handler = createHandler(context.GlobalString("assets"))
		)

		if manager, err = cluster.New(scheduler.NewResourceManager()); err != nil {
			logger.Fatal(err)
		}

		if err := http.ListenAndServe(context.GlobalString("addr"), handler); err != nil {
			logger.Fatal(err)
		}
	}

	if err := app.Run(os.Args); err != nil {
		logger.Fatal(err)
	}
}