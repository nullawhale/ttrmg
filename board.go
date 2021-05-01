package main

import (
	"encoding/json"
	"fmt"
	"github.com/fatih/color"
	"os"
	"strings"
	"unicode"
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

func (db *database) readFromDB() error {
	read, err := os.OpenFile(dbPath, os.O_RDONLY, 0666)
	if err != nil {
		return err
	}
	err = json.NewDecoder(read).Decode(&db)
	return err
}

func (db *database) writeToDB() error {
	write, err := os.OpenFile(dbPath, os.O_WRONLY, 0666)
	if err != nil {
		return err
	}
	data, err := json.Marshal(&db)
	if err != nil {
		return err
	}
	_, err = write.Write(data)
	return err
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

	err = db.writeToDB()
	return err
}

func (db *database) addTask(t *task, bName string) error {
	var err error
	//var b board
	var maxID int64 = 0

	// TODO: maybe it should be default(system) Board with default(system) name
	if len(db.Boards) == 0 {
		err = db.addBoard(&board{Name: "My board", Status: false})
	}

	for _, board := range db.Boards {
		if board.Name == bName {
			//b = *board
			for _, task := range board.Tasks {
				if task.ID > maxID {
					maxID = task.ID
				}
			}
			t.ID = maxID + 1
			board.Tasks = append(board.Tasks, t)
			err = db.writeToDB()
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

	for _, board := range db.Boards {
		if board.Name == bName {
			//b = *board
			for _, task := range board.Tasks {
				if task.ID == taskId {
					task.Status = true
					err = db.writeToDB()
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

	percent = done * 100 / (done + inProgress)
	return fmt.Sprintf(
		"%s%d%% of all tasks complete\n%s%s done | %s in progress\n",
		strings.Repeat(" ", indent/2), percent,
		strings.Repeat(" ", indent/2), green(done), purple(inProgress),
	)
}

func (db *database) printDB() {
	for _, board := range db.Boards {
		fmt.Printf("%s@%s\n", strings.Repeat(" ", indent/2), u(board.Name))
		for _, task := range board.Tasks {
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
		fmt.Println()
	}
	fmt.Print(db.stat())
}
