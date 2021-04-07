package main

import (
	"embed"
	"flag"
	"html/template"
	"io/fs"
	"log"
	"net/http"
	"strconv"

	"local/bookmarks/datastore"
	"local/bookmarks/server"
)

//go:embed templates
var templateFS embed.FS

//go:embed static
var staticFS embed.FS

//go:embed schema
var schemaFS embed.FS

func main() {
	config := parseArgs()

	templates, err := template.ParseFS(templateFS, "templates/*")
	if err != nil {
		log.Fatal(err)
	}
	static, err := fs.Sub(staticFS, "static")
	if err != nil {
		log.Fatal(err)
	}

	datastore, err := datastore.Connect(config.dbFile)
	if err != nil {
		log.Fatalf("opening database: %s", err)
	}

	n, err := datastore.RunMigrations(schemaFS)
	if err != nil {
		log.Fatalf("running migrations: %s", err)
	}
	if n > 0 {
		log.Printf("Ran %d migrations", n)
	}

	router := server.MakeRouter(templates, static, &datastore)
	log.Printf("Serving HTTP on port %d\n", config.port)
	log.Fatal(http.ListenAndServe(":"+strconv.Itoa(int(config.port)), router))
}

type config struct {
	port   uint
	dbFile string
}

func parseArgs() config {
	config := config{}
	flag.UintVar(&config.port, "port", 8080, "port to serve on")
	flag.StringVar(&config.dbFile, "db", "./bookmarks.db", "location of the bookmarks database")
	flag.Parse()
	return config
}
