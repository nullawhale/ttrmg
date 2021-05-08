package main

import (
	"fmt"
	"github.com/jessevdk/go-flags"
	"log"
	"os"
	"strconv"
)

var dbPath = ".db.tt"

var opts struct {
	Board     string `short:"b" long:"board" description:"Specify board name" required:"false"`
	Task      string `short:"t" long:"task" description:"Specify task name" required:"false"`
	CheckTask string `short:"c" long:"check" description:"Check task" required:"false"`
}

func main() {

	p := flags.NewParser(&opts, flags.Default)
	if _, err := p.Parse(); err != nil {
		if err.(*flags.Error).Type != flags.ErrHelp {
			log.Printf("[ERROR] cli error: %v", err)
		}
		os.Exit(0)
	}

	if len(os.Args) == 1 {
		if _, err := os.Stat(dbPath); err == nil {
			database := database{}
			err = database.readFromDB()
			if err != nil {
				fmt.Println("Unable to read db file:", err)
				os.Exit(1)
			}
			database.printDB()
		} else {
			file, err := os.Create(dbPath)
			if err != nil {
				fmt.Println("Unable to create db file:", err)
				os.Exit(1)
			}
			defer file.Close()
			fmt.Println(file.Name())
		}
	}

	_, err := flags.ParseArgs(&opts, os.Args)
	if err != nil {
		fmt.Println("Unable to parse args:", err)
		os.Exit(1)
	}

	/*fmt.Printf("Board: %s\n", opts.Board)
	fmt.Printf("Task: %s\n", opts.Task)
	fmt.Printf("CheckTask: %s\n", opts.CheckTask)*/

	if opts.Task != "" {
		if opts.Board == "" {
			fmt.Println("You must provide a name of board.")
			os.Exit(1)
		} else {
			database := database{}
			err := database.readFromDB()
			if err != nil {
				fmt.Println(err)
			}

			err = database.addTask(&task{
				ID:     0,
				Text:   opts.Task,
				Status: false,
			}, opts.Board)
			if err != nil {
				fmt.Println(err)
			}

			fmt.Printf("task \"%s\" is now added to your %s board.\n", opts.Task, opts.Board)
		}
	}

	if opts.CheckTask != "" {
		if opts.Board == "" {
			fmt.Println("You must provide a name of board.")
			os.Exit(1)
		} else {
			database := database{}
			err := database.readFromDB()
			if err != nil {
				fmt.Println(err)
			}

			id, err := strconv.Atoi(opts.CheckTask)
			if err != nil {
				fmt.Println(err)
			}
			err = database.checkTask(int64(id), opts.Board)
			if err != nil {
				fmt.Println(err)
			}

			fmt.Printf("task with id %d from board %s checked as done\n", int64(id), opts.Board)
		}
	}
}
