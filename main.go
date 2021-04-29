package main

import (
	"fmt"
	"github.com/urfave/cli"
	"log"
	"os"
	"strconv"
)

var dbPath = ".db.tt"

func main() {
	app := cli.NewApp()
	app.Name = "ttrmg"
	app.Usage = "Tasks for cli"
	app.Version = "0.0.1"

	app.Action = func(c *cli.Context) error {
		if _, err := os.Stat(dbPath); err == nil {
			database := database{}
			err = database.readFromDB()
			if err != nil {
				return err
			}
			database.printDB()
		} else {
			file, err := os.Create(dbPath)
			if err != nil {
				fmt.Println("Unable to create file:", err)
				os.Exit(1)
			}
			defer file.Close()
			fmt.Println(file.Name())
		}
		return nil
	}

	app.Commands = []cli.Command{
		{
			Name:      "task",
			ShortName: "t",
			Usage:     "Create task",
			Action: func(c *cli.Context) {
				if len(c.Args()) != 2 {
					fmt.Println()
					fmt.Println("Error")
					fmt.Println("You must provide a name of your task and name of board.")
					fmt.Println("Example: ttrmg task boardName \"task text\"")
					fmt.Println()
					return
				}

				database := database{}
				err := database.readFromDB()
				if err != nil {
					fmt.Println(err)
				}

				err = database.addTask(&task{
					ID:     0,
					Text:   c.Args()[1],
					Status: false,
				}, c.Args()[0])
				if err != nil {
					fmt.Println(err)
				}

				fmt.Printf("task \"%s\" is now added to your %s board.\n", c.Args()[1], c.Args()[0])
			},
		},
		{
			Name:      "check",
			ShortName: "c",
			Usage:     "Check task",
			Action: func(c *cli.Context) {
				if len(c.Args()) != 2 {
					fmt.Println()
					fmt.Println("Error")
					fmt.Println("You must provide a name of your task and name of board.")
					fmt.Println("Example: ttrmg task boardName 2")
					fmt.Println("Example: ttrmg task boardName 5")
					fmt.Println()
					return
				}

				database := database{}
				err := database.readFromDB()
				if err != nil {
					fmt.Println(err)
				}

				id, err := strconv.Atoi(c.Args()[1])
				if err != nil {
					fmt.Println(err)
				}
				err = database.checkTask(int64(id), c.Args()[0])
				if err != nil {
					fmt.Println(err)
				}

				fmt.Printf("task with id %d from board %s checked as done\n", int64(id), c.Args()[0])
			},
		},
	}

	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}
