package main

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/jack-zh/gotodo/task"
)

var noAct = errors.New("didn't act")

var (
	file = flag.String("file", defaultFile(".todo", "TODO"), "file in which to store tasks")
	log  = flag.String("log", defaultFile(".todo-log", "TODOLOG"), "file in which to log removed tasks")
	now  = flag.Bool("now", false, "when adding, insert at head")
	done = flag.Bool("done", false, "don't actually add; just append to log file")
)

func defaultFile(name, env string) string {
	if f := os.Getenv(env); f != "" {
		return f
	}
	return filepath.Join(os.Getenv("HOME"), name)
}

const usage = `Usage:
	todo
		Show top task
	todo ls
		Show all tasks
	todo N
		Promote task N to top
	todo rm
		Remove top task
	todo rm N
		Remove task N
Flags:
`

func main() {
	flag.Usage = func() {
		fmt.Fprint(os.Stderr, usage)
		flag.PrintDefaults()
	}
	flag.Parse()

	list := task.NewList(*file)
	a, n := flag.Arg(0), len(flag.Args())

	err := noAct
	switch {
	case a == "ls" && n == 1:
		// list tasks
		var tasks []string
		tasks, err = list.Get()
		for i := len(tasks) - 1; i >= 0; i-- {
			fmt.Printf("%2d: %s\n", i, tasks[i])
		}
	case a == "rm" && n <= 2:
		i, err2 := strconv.Atoi(flag.Arg(1))
		if err2 != nil && n == 2 {
			break
		}
		if n == 1 {
			i = 0
		}
		var t string
		t, err = list.GetTask(i)
		if err != nil {
			break
		}
		if err2 := logRemovedTask(t); err2 != nil {
			fmt.Fprintln(os.Stderr, "logging:", err2)
		}
		err = list.RemoveTask(i)
		if err != nil || n == 2 {
			break
		}
		fallthrough
	case n == 0:
		var tasks []string
		tasks, err = list.Get()
		if len(tasks) > 0 {
			fmt.Println(tasks[0])
		}
	case n == 1:
		i, err2 := strconv.Atoi(flag.Arg(0))
		if err2 != nil {
			break
		}
		var t string
		t, err = list.GetTask(i)
		if err != nil {
			break
		}
		err = list.RemoveTask(i)
		if err != nil {
			break
		}
		err = list.AddTask(t, true)
		if err == nil {
			fmt.Println(t)
		}
	}
	if err == noAct {
		t := strings.Join(flag.Args(), " ")
		if *done {
			err = logRemovedTask(t)
		} else {
			err = list.AddTask(t, *now)
		}
	}
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func logRemovedTask(t string) error {
	f, err := os.OpenFile(*log, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0644)
	if err != nil {
		return err
	}
	defer f.Close()
	now := time.Now().Format("1986-02-17")
	_, err = fmt.Fprintf(f, "%s %s\n", now, t)
	return err
}
