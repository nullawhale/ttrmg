package main

import (
	"fmt"
	"github.com/jessevdk/go-flags"
	"os"
	"strings"
)

type Options struct {
	DbPath string   `long:"db-path" description:"Database path" default:".db.tt"`
	New    struct{} `command:"new" alias:"todo" alias:"n" alias:"make" description:"Add new task" required:"false"`
	List   struct{} `command:"list" alias:"l" alias:"ls" description:"List tasks" required:"false"`
	Done   struct{} `command:"done" alias:"d" description:"Check task as done" required:"false"`
}

var options Options

var parser = flags.NewParser(&options, flags.Default)

func main() {
	parser.Command.SubcommandsOptional = true
	args, err := parser.Parse()
	if err != nil {
		if err.(*flags.Error).Type == flags.ErrHelp {
			os.Exit(0)
		}
		parser.WriteHelp(os.Stderr)
		os.Exit(1)
	}

	db, err := ReadDatabaseFromFile(options.DbPath)
	if err != nil {
		if os.IsNotExist(err) {
			db = NewDatabase()
		} else {
			panic(err)
		}
	}
	db, err = db.reCalcTasks()
	if err != nil {
		fmt.Println(err.Error())
	}
	defer db.WriteToFile(options.DbPath)

	var command string
	if parser.Active != nil {
		command = parser.Active.Name
	} else if len(args) == 0 {
		command = "list"
	} else {
		command = "new"
	}
	switch command {
	case "new":
		err = db.NewTask(strings.Join(args, " "))
		if err != nil {
			fmt.Println(err.Error())
		}
	case "list":
		db.printDB(strings.Join(args, " "))
	case "done":
		err = db.checkTask(strings.Join(args, " "))
		if err != nil {
			fmt.Println(err.Error())
		}
	}
}

// vi:noet:ts=4:sw=4:
