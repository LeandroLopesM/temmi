package main

import (
	"flag"
	"fmt"
	"os"

	"charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"
	"github.com/charmbracelet/log"
	"github.com/leandrolopesm/eunuch/engine"
	eunuch "github.com/leandrolopesm/eunuch/lib"
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

var windowSize tea.WindowSizeMsg

type Choice struct {
	value  string
	target string
	hook   util.Option[eunuch.Scheme]
}

type Passage struct {
	hook    util.Option[eunuch.Scheme]
	text    string
	choices []Choice
}

type model struct {
	backend textEngine
	ready bool
}

func initState(file, passage string) model {
	m := model{
		backend: textEngine {
			engine:      engine.New(),
			passages:    make(map[string]Passage),
			currPassage: passage,
			storyFile:   file,

			lineReader: readline {
				reader: textinput.New(),
				active: false,
			},
		},
		ready: false,
	};

	m.backend.lineReader.reader.CharLimit = 156

	m.backend.LoadStory()
	return m
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		if !m.ready {m.ready = true}
		windowSize = msg

		m.backend.lineReader.reader.SetWidth(msg.Width)
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

	var cmd tea.Cmd
	m.backend.lineReader.reader, cmd = m.backend.lineReader.reader.Update(msg)
	return m, cmd
}

func (m model) View() tea.View {
	var str string
	if !m.ready {
		str = "Initializing...\n"
	} else {
		str = m.backend.View()
	}

	return tea.NewView(str)
}

func main() {
	f := util.Assert(os.OpenFile("stderr.log", os.O_WRONLY|os.O_CREATE, 0644))
	defer f.Close()

	log.SetReportTimestamp(false)
	log.SetLevel(log.DebugLevel)
	log.SetOutput(f)

	var storyFile = flag.String("file", "default-story.scm", "Story file")
	var passage = flag.String("start", "start", "Passage to start from")
	var help = flag.Bool("help", false, "Help message")

	flag.Parse()

	if *help {
		flag.Usage()
		return
	}

	m := initState(*storyFile, *passage)
	p := tea.NewProgram(m)

	if _, err := p.Run(); err != nil {
		fmt.Printf("Alas, there's been an error: %v", err)
	}
}
