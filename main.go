package main

import (
	"embed"
	"flag"
	"fmt"
	"io/fs"
	"local/bookmarks/datastore"
	"local/bookmarks/server"
	"local/bookmarks/templates"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
)

//go:embed pages
var templateFS embed.FS

//go:embed static
var staticFS embed.FS

//go:embed schema
var schemaFS embed.FS

var commandList []command

type command struct {
	flags *flag.FlagSet
	run   func()
}

func main() {
	commandList = []command{
		serverCommand(),
		manageUserCommand(),
		helpCommand(),
	}
	if len(os.Args) < 2 {
		printHelp()
		os.Exit(1)
	}
	commandName := os.Args[1]
	for _, command := range commandList {
		if command.flags.Name() == commandName {
			command.run()
			return
		}
	}
	fmt.Printf("unknown command \"%s\"\n", commandName)
	os.Exit(1)
}

type serveConfig struct {
	port   uint
	dbFile string
}

func serverCommand() command {
	config := serveConfig{}
	flags := flag.NewFlagSet("serve", flag.ContinueOnError)
	flags.UintVar(&config.port, "port", 8080, "port to serve on")
	flags.StringVar(&config.dbFile, "db", "./bookmarks.db", "location of the bookmarks database")
	return command{
		flags: flags,
		run: func() {
			flags.Parse(os.Args[2:])
			serve(config)
		},
	}
}

func serve(config serveConfig) {
	templates := templates.CreateTemplates(templateFS)

	static, err := fs.Sub(staticFS, "static")
	if err != nil {
		panic(err)
	}

	ds, err := openDatabase(config.dbFile)
	if err != nil {
		log.Fatalf("opening database file %s: %s", config.dbFile, err)
	}

	ds.CleanUpCookies(time.Hour * 24 * 30)

	router := server.MakeRouter(&templates, static, ds)
	log.Printf("Serving HTTP on port %d\n", config.port)
	log.Fatal(http.ListenAndServe(":"+strconv.Itoa(int(config.port)), router))
}

type manageUserConfig struct {
	username  string
	password  string
	delete    bool
	listUsers bool
	dbFile    string
}

func manageUserCommand() command {
	config := manageUserConfig{}
	flags := flag.NewFlagSet("user", flag.ContinueOnError)
	flags.StringVar(&config.username, "username", "", "Username to update")
	flags.StringVar(&config.password, "password", "", "Password to set")
	flags.BoolVar(&config.delete, "delete", false, "Delete this user instead of updating it")
	flags.BoolVar(&config.listUsers, "list", false, "List all users, then exit")
	flags.StringVar(&config.dbFile, "db", "./bookmarks.db", "location of the bookmarks database")
	return command{
		flags: flags,
		run: func() {
			flags.Parse(os.Args[2:])
			manageUser(config)
		},
	}
}

func manageUser(config manageUserConfig) {
	ds, err := openDatabase(config.dbFile)
	if err != nil {
		fmt.Printf("opening database file %s: %s\n", config.dbFile, err)
		os.Exit(1)
	}

	if config.listUsers {
		list, err := ds.ListUsers()
		if err != nil {
			fmt.Printf("getting users: %s\n", err)
			os.Exit(1)
		}
		fmt.Println("All users:")
		for _, user := range list {
			fmt.Printf(" %s\n", user)
		}
		return
	}

	if config.username != "" {
		if config.delete {
			err = ds.RemoveUser(config.username)
			if err != nil {
				fmt.Printf("removing user %s: %s\n", config.username, err)
				os.Exit(1)
			}
			fmt.Printf("Removed user %s\n", config.username)
		} else {
			if config.password != "" {
				_, exists, err := ds.UserExists(config.username)
				if err != nil {
					fmt.Printf("checking whether user exists: %s\n", err)
					os.Exit(1)
				}
				if exists {
					err = ds.ChangeUserPassword(config.username, config.password)
					if err != nil {
						fmt.Printf("changing user %s's password: %s\n", config.username, err)
						os.Exit(1)
					}
					fmt.Printf("Changed %s's password\n", config.username)
				} else {
					err = ds.AddUser(config.username, config.password)
					if err != nil {
						fmt.Printf("adding user %s: %s\n", config.username, err)
						os.Exit(1)
					}
					fmt.Printf("Added user %s\n", config.username)
				}
			} else {
				fmt.Printf("To create a user or change a user's password, password must be non-empty\n")
				os.Exit(1)
			}
		}
	} else {
		fmt.Printf("Username must be non-empty\n")
		os.Exit(1)
	}
}

func helpCommand() command {
	flags := flag.NewFlagSet("help", flag.ContinueOnError)
	return command{
		flags: flags,
		run: func() {
			printHelp()
		},
	}
}

func openDatabase(dbFile string) (*datastore.Datastore, error) {
	datastore, err := datastore.Connect(dbFile)
	if err != nil {
		return nil, fmt.Errorf("opening database: %w", err)
	}

	n, err := datastore.RunMigrations(schemaFS)
	if err != nil {
		return nil, fmt.Errorf("running migrations: %w", err)
	}
	if n > 0 {
		log.Printf("Ran %d migrations\n", n)
	}
	return &datastore, nil
}

func printHelp() {
	commandNames := make([]string, 0, len(commandList))
	for _, command := range commandList {
		commandNames = append(commandNames, command.flags.Name())
	}
	fmt.Printf("Usage: %s [ COMMAND ] [ FLAGS ]\nCommands: %s\n", os.Args[0], strings.Join(commandNames, ", "))
}
