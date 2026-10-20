package main

import (
	"errors"
	"flag"
	"fmt"
	"os"

	tea "charm.land/bubbletea/v2"
	"github.com/charmbracelet/log"
	"github.com/leandrolopesm/eunuch/core"
	"github.com/leandrolopesm/eunuch/engine"
	"github.com/leandrolopesm/eunuch/util"
)

var (
	/// Should update model.windowSize
	optResponsive bool = true
	/// Should break passage text when printing
	optWrapLines bool = true
	/// Fixed line width, won't have effect unless optResponsive == false
	optLineWidth int = 50
)

var (
	glbReady = false
)

var windowSize tea.WindowSizeMsg

type Choice struct {
	value  string
	target string
	hook   core.Value
}

type Passage struct {
	hook    core.Value
	text    string
	choices []Choice
}

type model struct {
	backend textEngine
}

func initState(file string) model {
	m := model{
		backend: textEngine {
			engine:      engine.New(),
			passages:    make(map[string][]core.Value),
			currPassage: Passage{},
			storyFile:   file,
		},
	};

	m.backend.LoadStory()
	return m
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		if !glbReady {
			if optResponsive {
				optLineWidth = msg.Width
			}

			glbReady = true
		}
	case tea.KeyPressMsg:
		switch msg.String() {
		case "ctrl+c", "ctrl+d", "q":
			return m, tea.Quit

		case "up":
			m.backend.choiceUp()
		case "down":
			m.backend.choiceDown()
		case "enter", "space":
			m.backend.commitChoice()
		}
	}

	return m, nil
}

func (m model) View() tea.View {
	str := m.backend.View()

	return tea.NewView(str)
}

func main() {
	var storyFile = flag.String("file", "NONE", "Story file")
	var help = flag.Bool("help", false, "Help message")

	flag.Parse()

	if *help || *storyFile == "NONE" {
		flag.Usage()
		return
	}

	f := openLogFile(*storyFile)
	defer f.Close()

	log.SetReportTimestamp(false)
	log.SetLevel(log.DebugLevel)
	log.SetOutput(f)


	m := initState(*storyFile)
	p := tea.NewProgram(m)

	if _, err := p.Run(); err != nil {
		fmt.Printf("There's been an error: %v", err)
	}
}

func openLogFile(storyName string) *os.File {
	var i = 0
	for {
		fileName := fmt.Sprintf("%s-%d.log", storyName, i)
		if _,err := os.Stat(fileName); err == nil {
			i++
		} else if errors.Is(err, os.ErrNotExist) {
			return util.Assert(os.OpenFile(fileName, os.O_WRONLY|os.O_CREATE, 0644))
		}
	}
}