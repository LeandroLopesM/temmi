package main

import (
	"fmt"
	"os"

	"github.com/charmbracelet/log"
	"github.com/kr/pretty"
	"github.com/leandrolopesm/eunuch/builtin"
	"github.com/leandrolopesm/eunuch/lib"

	"charm.land/bubbles/v2/textinput"
)

type readline struct {
    reader textinput.Model
    active bool
}

type textEngine struct {
	engine   lib.Engine
	passages map[string]([]lib.Scheme)

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

		log.Infof("Finalized map: %# v", pretty.Formatter(m.passages))

		if pass, ok := m.passages["start"]; !ok {
			log.Fatal("Story file does not specify a 'start' passage")
		} else {
			log.Infof("Start passage: %# v", pretty.Formatter(pass))
		}
	}

	m.passages["__unknown_passage"] = m.engine. "(passage \"__unknown_passage\" (text (red \"ERROR\") \": That choice linked to a non-existent passage\nSee the log file for more details\"))"
}

func (t *textEngine) View() string {
    passage,_ := t.passages[t.currPassage]
	str := passage.text

    if len(str) == 0 {
		str += "This passage is empty"
	}

	str += "\n\n"

	for i, choice := range passage.choices {
		cursor := " "
		if t.choiceIdx == i {
			cursor = "-> "
		}

		str += fmt.Sprintf("%s%s\n", cursor, choice.value)
	}
    str += "\nPress 'q' to exit."

    return tryWrapLines(str)
}

func (m *textEngine) loadPassage(passage string) Passage {
	var output Passage
	PassageBuilderTarget = &output

	if program, ok := m.passages[passage]; !ok {
		log.Errorf("Undefined passage %v", passage)
		m.currPassage = m.loadPassage("__unknown_passage")
	} else {
		for _,scm := range program {
			m.engine.Evaluate(scm)
		}
	}

	PassageBuilderTarget = nil
	return output
}

func (m *textEngine) commitChoice() {
	if len(m.currPassage.choices) != 0 {
        choice := m.currPassage.choices[m.choiceIdx]
        m.callChoiceHook(choice)

        last := m.currPassage
		m.currPassage = m.loadPassage(choice.target)

		m.choiceIdx = 0
        m.callOnLoadHook(last)
    }
}

func (m *textEngine) choiceUp() {
    if m.choiceIdx-1 < 0 {
        m.choiceIdx = len(m.passages[m.currPassage].choices) - 1
    } else {
        m.choiceIdx--
    }
}

func (m *textEngine) choiceDown() {
    if m.choiceIdx+1 > len(m.passages[m.currPassage].choices) {
        m.choiceIdx = 0
    } else {
        m.choiceIdx++
    }
}

func (m *textEngine) callChoiceHook(choice Choice) {
	if hook,err := choice.hook.Try(); err == nil {
		if err := m.engine.Evaluate(hook); err != nil {
			log.Errorf("Passage %s: Failed run hook for choice '%s': %v", m.currPassage, choice.value, err)
		}
	}
}
func (m *textEngine) callOnLoadHook(lastPassage Passage) {
	if hook,err := m.currPassage.hook.Try(); err == nil {
		if err := m.engine.Evaluate(hook); err != nil {
			log.Errorf("Failed to load '%s': %v", m.currPassage, err)
			m.currPassage = lastPassage
		}
	}
}