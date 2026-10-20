package main

import (
	"fmt"
	"os"

	"github.com/charmbracelet/log"
	"github.com/kr/pretty"
	"github.com/leandrolopesm/eunuch/builtin"
	"github.com/leandrolopesm/eunuch/core"
	"github.com/leandrolopesm/eunuch/lib"
)

type textEngine struct {
	engine   lib.Engine
	passages map[string]([]core.Value)

	currPassage Passage
	choiceIdx   int

	storyFile string
}

func (m *textEngine) LoadStory() {
	builtin.RegisterSelf(&m.engine)
	Register(m)

	if fileStr, err := os.ReadFile(m.storyFile); err != nil {
		fmt.Printf("Failed to open story '%v' (%v)", m.storyFile, err)
		os.Exit(1)
	} else {
		if err := m.engine.ExecuteStr(m.storyFile, string(fileStr)); err != nil {
			fmt.Printf("Failed to load story '%v':\n%v", m.storyFile, err)
			os.Exit(1)
		}

		log.Infof("Final story map: %# v", pretty.Formatter(m.passages))

		if pass, ok := m.passages["start"]; !ok {
			log.Fatal("Story file does not specify a 'start' passage")
		} else {
			log.Infof("Start passage: %# v", pretty.Formatter(pass))
			m.loadPassage("start")
		}
	}

	rawValues,_ := m.engine.CompileStr("engine", "(passage \"__unknown_passage\" (text (red \"ERROR\") \": That choice linked to a non-existent passage\\nSee the log file for more details\"))")
	m.passages["__unknown_passage"] = rawValues
}

func (t *textEngine) View() string {
	var str string
	if glbReady {
		str = t.currPassage.text

		if len(str) == 0 {
			str += "This passage is empty "
		}

		str += "\n\n"

		for i, choice := range t.currPassage.choices {
			cursor := " "
			if t.choiceIdx == i {
				cursor = "-> "
			}

			str += fmt.Sprintf("%s%s\n", cursor, choice.value)
		}

		str += "\nPress 'q' to exit."
		return tryWrapLines(str)
	} else {
		return "Initializing...\n"
	}
}

func (m *textEngine) loadPassage(passage string) {
	// Update global state to match this instance
	// since bubbletea doesn't pass by pointer
	Model = m
	m.currPassage = Passage{}

	if program, ok := m.passages[passage]; !ok {
		log.Errorf("Undefined passage %v", passage)
		m.loadPassage("__unknown_passage")
	} else {
		for _,scm := range program {
			log.Debugf("Eval %s: %s", scm.TypeString(), scm.ToString())
			if err := m.engine.Evaluate(scm); err != nil {
				log.Errorf("Failed to load passage '%s': %s", passage, err)
			}
		}
	}
}

func (m *textEngine) commitChoice() {
	if len(m.currPassage.choices) != 0 {
        choice := m.currPassage.choices[m.choiceIdx]
        m.callChoiceHook(choice)

		m.loadPassage(choice.target)
		m.choiceIdx = 0
    }
}

func (m *textEngine) choiceUp() {
    if m.choiceIdx-1 < 0 {
        m.choiceIdx = len(m.currPassage.choices) - 1
    } else {
        m.choiceIdx--
    }
}

func (m *textEngine) choiceDown() {
    if m.choiceIdx+1 > len(m.currPassage.choices) {
        m.choiceIdx = 0
    } else {
        m.choiceIdx++
    }
}

func (m *textEngine) callChoiceHook(choice Choice) {
	if choice.hook != nil {
		if err := m.engine.Evaluate(choice.hook); err != nil {
			log.Errorf("Passage %s: Failed run hook for choice '%s': %v", m.currPassage, choice.value, err)
		}
	}
}