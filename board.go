package main

import (
	"encoding/json"
	"fmt"
	"github.com/manifoldco/promptui"
	"os"
	"strings"
	"time"
	"unicode"

	"github.com/fatih/color"
	"github.com/lithammer/fuzzysearch/fuzzy"
)

type Database struct {
	Boards []*Board `json:"boards"`
}

type Board struct {
	Name   string  `json:"name"`
	Status bool    `json:"status"`
	Tasks  []*Task `json:"tasks"`
}

type Task struct {
	ID     int64  `json:"id"`
	Text   string `json:"name"`
	Status bool   `json:"status"`
	Date   string `json:"date"`
}

var green = color.New(color.FgGreen).SprintFunc()
var purple = color.New(color.FgMagenta).SprintFunc()
var gray = color.New(color.FgHiBlack).SprintFunc()
var u = color.New(color.Underline).SprintFunc()

const (
	TaskActual string = "actual"
	TaskMonth  string = "month"
	TaskRotten string = "rotten"
)

var indent = 10

func NewDatabase() *Database {
	return &Database{[]*Board{}}
}

func ReadDatabaseFromFile(name string) (*Database, error) {
	file, err := os.Open(name)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var db Database
	err = json.NewDecoder(file).Decode(&db)
	return &db, err
}

func (db *Database) WriteToFile(name string) error {
	file, err := os.OpenFile(name, os.O_WRONLY|os.O_CREATE, 0600)
	if err != nil {
		return err
	}
	defer file.Close()

	return json.NewEncoder(file).Encode(db)
}

func (db *Database) addBoard(b *Board) error {
	var err error

	if len(b.Name) >= 33 {
		return fmt.Errorf("board name must not be longer than 32 characters")
	}
	for _, v := range b.Name {
		if unicode.IsSpace(v) {
			return fmt.Errorf("board name must not contain spaces")
		} else if unicode.IsPunct(v) {
			return fmt.Errorf("board name must not contain special characters")
		}
	}

	for _, board := range db.Boards {
		if board.Name == b.Name {
			return nil
		}
	}

	db.Boards = append(db.Boards, b)

	return err
}

func (db *Database) NewTask(text string) error {
	dateNow := time.Now()
	err := db.addTask(
		&Task{Text: text, Date: dateNow.Format(time.DateTime)},
		TaskActual,
	)
	if err != nil {
		return err
	}

	return err
}

func (db *Database) addTask(t *Task, bName string) error {
	var err error
	var maxID int64 = 0

	if bName == "" {
		bName = TaskActual
	}

	if len(db.Boards) == 0 {
		err := db.addBoard(&Board{Name: bName, Status: false})
		if err != nil {
			return err
		}
	}

	for _, board := range db.Boards {
		if strings.ToLower(board.Name) == strings.ToLower(bName) {
			for _, task := range board.Tasks {
				if task.Text == t.Text {
					fmt.Println("Task already exists")
					return err
				}
				if task.ID > maxID {
					maxID = task.ID
				}
			}
			t.ID = maxID + 1
			board.Tasks = append(board.Tasks, t)
			return err
		}
	}

	return err
}

func (db *Database) checkTask(taskPattern string) error {
	var err error
	var matchedTasks []*Task

	for _, board := range db.Boards {
		for _, task := range board.Tasks {
			if task.Status != true {
				if fuzzy.MatchFold(taskPattern, task.Text) && fuzzy.RankMatch(taskPattern, task.Text) >= 0 {
					matchedTasks = append(matchedTasks, task)
				}
			}
		}
	}

	if matchedTasks != nil {
		if len(matchedTasks) == 1 {
			matchedTasks[0].Status = true
			db.printDB("")
		} else {
			var s []string
			var foundTaskString string
			for _, task := range matchedTasks {
				//fmt.Printf("%s\n", Task.Text)
				s = append(s, task.Text)
			}
			prompt := promptui.Select{
				Label: "Found more than one Task. Select one:",
				Items: s,
			}

			_, foundTaskString, err = prompt.Run()
			if err != nil {
				return fmt.Errorf("Prompt failed %v\n", err)
			}

			for _, board := range db.Boards {
				for _, task := range board.Tasks {
					if foundTaskString == task.Text {
						task.Status = true
						db.printDB("")
					}
				}
			}
		}
		//fmt.Printf("Task \"%s\" checked as done\n", matchedTask.Text)
	} else {
		return fmt.Errorf("task not found")
	}

	return err
}

func (db *Database) stat() string {
	var done int64 = 0
	var inProgress int64 = 0
	var percent int64 = 0

	for _, board := range db.Boards {
		if len(board.Tasks) != 0 {
			for _, task := range board.Tasks {
				if task.Status {
					done++
				} else {
					inProgress++
				}
			}
		}
	}
	if done+inProgress == 0 {
		return strings.Repeat(" ", indent/2) + "No tasks were found.\n"
	}
	percent = done * 100 / (done + inProgress)
	return fmt.Sprintf(
		"%s%d%% of all tasks complete\n%s%s done | %s in progress\n",
		strings.Repeat(" ", indent/2), percent,
		strings.Repeat(" ", indent/2), green(done), purple(inProgress),
	)
}

func (db *Database) printDB(pattern string) {
	for _, board := range db.Boards {
		fmt.Printf("%s@%s\n", strings.Repeat(" ", indent/2), u(board.Name))
		for _, task := range board.Tasks {
			if fuzzy.MatchFold(pattern, task.Text) {
				var id = fmt.Sprintf("%d.", task.ID)
				if task.Status {
					fmt.Printf("%s%s %s %s\n",
						strings.Repeat(" ", indent-len(id)),
						gray(id), green("[✓]"), gray(task.Text))
				} else {
					fmt.Printf("%s%s %s %s %s\n",
						strings.Repeat(" ", indent-len(id)),
						gray(id), purple("[ ]"), task.Text, task.Date)
				}
			}
		}
		fmt.Println()
	}
	fmt.Print(db.stat())
}

func (db *Database) reCalcTasks() (*Database, error) {
	var tmpDb = NewDatabase()
	for _, board := range db.Boards {
		//fmt.Println(board.Name)
		for _, task := range board.Tasks {
			taskDate, err := time.Parse(time.DateTime, task.Date)
			if err != nil {
				return nil, fmt.Errorf("wrong date format")
			}
			now := time.Now()
			duration := now.Sub(taskDate)
			hours := int(duration.Hours())

			if hours < 24*7 {
				// started no later than a week
				err = tmpDb.addBoard(&Board{Name: TaskActual, Status: false})
				err = tmpDb.addTask(task, TaskActual)
				if err != nil {
					return nil, fmt.Errorf(err.Error())
				}
			} else if hours > 24*7 && hours < 24*30 {
				// started no later than a month
				err = tmpDb.addBoard(&Board{Name: TaskMonth, Status: false})
				err = tmpDb.addTask(task, TaskMonth)
				if err != nil {
					return nil, fmt.Errorf(err.Error())
				}
			} else if hours > 24*30 {
				// started from month and later
				err = tmpDb.addBoard(&Board{Name: TaskRotten, Status: false})
				err = tmpDb.addTask(task, TaskRotten)
				if err != nil {
					return nil, fmt.Errorf(err.Error())
				}
			}
		}
	}

	return tmpDb, nil
}

// vi:noet:ts=4:sw=4:
