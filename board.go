package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"unicode"

	"github.com/fatih/color"
	"github.com/lithammer/fuzzysearch/fuzzy"
)

type database struct {
	Boards []*board `json:"boards"`
}

type board struct {
	ID     int64   `json:"id"`
	Name   string  `json:"name"`
	Status bool    `json:"status"`
	Tasks  []*task `json:"tasks"`
}

type task struct {
	ID     int64  `json:"id"`
	Text   string `json:"name"`
	Status bool   `json:"status"`
}

var green = color.New(color.FgGreen).SprintFunc()
var purple = color.New(color.FgMagenta).SprintFunc()
var gray = color.New(color.FgHiBlack).SprintFunc()
var u = color.New(color.Underline).SprintFunc()

var indent = 10

func NewDatabase() *database {
	return &database{[]*board{}}
}

func ReadDatabaseFromFile(name string) (*database, error) {
	file, err := os.Open(name)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var db database
	err = json.NewDecoder(file).Decode(&db)
	return &db, err
}

func (db *database) WriteToFile(name string) error {
	file, err := os.OpenFile(name, os.O_WRONLY|os.O_CREATE, 0600)
	if err != nil {
		return err
	}
	defer file.Close()

	return json.NewEncoder(file).Encode(db)
}

func (db *database) addBoard(b *board) error {
	var err error
	var maxID int64 = 0

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
			return fmt.Errorf("board with this name already exist")
		}
		if board.ID > maxID {
			maxID = board.ID
		}
	}

	b.ID = maxID + 1
	db.Boards = append(db.Boards, b)

	return err
}

func (db *database) NewTask(text string) {
	db.addTask(&task{Text: text}, "actual")
}

func (db *database) addTask(t *task, bName string) error {
	var err error
	var maxID int64 = 0

	// TODO: maybe it should be default(system) Board with default(system) name
	if len(db.Boards) == 0 {
		err = db.addBoard(&board{Name: bName, Status: false})
	}

	for _, board := range db.Boards {
		if strings.ToLower(board.Name) == strings.ToLower(bName) {
			for _, task := range board.Tasks {
				if task.ID > maxID {
					maxID = task.ID
				}
			}
			t.ID = maxID + 1
			board.Tasks = append(board.Tasks, t)
			return err
		}
	}

	// TODO: what if there's no Boards with given name: print error or add first one
	/*for _, task := range db.Boards[0].Tasks {
		if task.ID > maxID {
			maxID = task.ID
		}
	}
	t.ID = maxID + 1
	db.Boards[0].Tasks = append(db.Boards[0].Tasks, t)
	err = db.writeToDB()
	return err*/

	return err
}

func (db *database) checkTask(taskId int64, bName string) error {
	var err error
	bName = strings.ToLower(bName)

	for _, board := range db.Boards {
		if strings.ToLower(board.Name) == strings.ToLower(bName) {
			for _, task := range board.Tasks {
				if task.ID == taskId {
					task.Status = true
					return err
				}
			}
		}
	}

	return err
}

func (db *database) stat() string {
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

func (db *database) printDB(pattern string) {
	for _, board := range db.Boards {
		fmt.Printf("%s@%s\n", strings.Repeat(" ", indent/2), u(board.Name))
		for _, task := range board.Tasks {
			if fuzzy.Match(pattern, task.Text) {
				var id = fmt.Sprintf("%d.", task.ID)
				if task.Status {
					fmt.Printf("%s%s %s %s\n",
						strings.Repeat(" ", indent-len(id)),
						gray(id), green("[✓]"), gray(task.Text))
				} else {
					fmt.Printf("%s%s %s %s\n",
						strings.Repeat(" ", indent-len(id)),
						gray(id), purple("[ ]"), task.Text)
				}
			}
		}
		fmt.Println()
	}
	fmt.Print(db.stat())
}

// vi:noet:ts=4:sw=4:
