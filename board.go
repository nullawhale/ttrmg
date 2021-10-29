package main

import (
	"fmt"
	"github.com/manifoldco/promptui"
	"strings"
	"unicode"

	"github.com/fatih/color"
	"github.com/lithammer/fuzzysearch/fuzzy"
)

var green = color.New(color.FgGreen).SprintFunc()
var purple = color.New(color.FgMagenta).SprintFunc()
var gray = color.New(color.FgHiBlack).SprintFunc()
var u = color.New(color.Underline).SprintFunc()

var indent = 10

func (db *Database) addBoard(b *Board) error {
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

func (db *Database) addTask(text string, bName string) error {
	var err error
	var task *Task
	var maxID int64 = 0

	task = &Task{Text: text}
	if err != nil {
		return err
	}

	// TODO: maybe it should be default(system) Board with default(system) name
	if len(db.Boards) == 0 {
		err := db.addBoard(&Board{Name: bName, Status: false})
		if err != nil {
			return err
		}
	}

	for _, board := range db.Boards {
		if strings.ToLower(board.Name) == strings.ToLower(bName) {
			for _, task := range board.Tasks {
				if task.Text == task.Text {
					fmt.Println("task already exists")
					return err
				}
				if task.ID > maxID {
					maxID = task.ID
				}
			}
			task.ID = maxID + 1
			board.Tasks = append(board.Tasks, task)
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

	if matchedTasks == nil {
		return fmt.Errorf("task not found")
	}

	if len(matchedTasks) == 1 {
		matchedTasks[0].Status = true
		db.printDB("")
	} else {
		var s []string
		var foundTaskString string
		for _, task := range matchedTasks {
			//fmt.Printf("%s\n", task.Text)
			s = append(s, task.Text)
		}
		prompt := promptui.Select{
			Label: "Found more than one task. Select one:",
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
	//fmt.Printf("task \"%s\" checked as done\n", matchedTask.Text)

	return err
}

func (db *Database) rmTask(taskPattern string) error {
	var err error

	type ResTask struct {
		Text  string
		Board string
		Index int
	}

	var matchedTasks []ResTask

	for _, board := range db.Boards {
		for i, task := range board.Tasks {
			if fuzzy.MatchFold(taskPattern, task.Text) && fuzzy.RankMatch(taskPattern, task.Text) >= 0 {
				matchedTasks = append(matchedTasks, ResTask{task.Text, board.Name, i})
			}
		}
	}

	if matchedTasks == nil {
		return fmt.Errorf("no matched task found")
	}

	if len(matchedTasks) == 1 {
		db.deleteTask(matchedTasks[0].Board, matchedTasks[0].Index)
		db.printDB("")
	} else {
		var s []string
		var foundTaskString string
		for _, task := range matchedTasks {
			s = append(s, task.Text)
		}
		prompt := promptui.Select{
			Label: "Found more than one task. Select one:",
			Items: s,
		}

		_, foundTaskString, err = prompt.Run()
		if err != nil {
			return fmt.Errorf("Prompt failed %v\n", err)
		}

		for i, task := range matchedTasks {
			if foundTaskString == task.Text {
				db.deleteTask(task.Board, i)
				db.printDB("")
			}
		}
	}

	return err
}

func (db *Database) deleteTask(boardName string, i int) {
	for _, board := range db.Boards {
		if boardName == board.Name {
			if len(board.Tasks) != 0 {
				board.Tasks = append(board.Tasks[:i], board.Tasks[i+1:]...)
			}
		}
	}
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
