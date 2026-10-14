package main

import (
	"fmt"
	"os"

	"github.com/charmbracelet/log"
	"github.com/kr/pretty"
	"github.com/leandrolopesm/eunuch/builtin"
	"github.com/leandrolopesm/eunuch/core"
	"github.com/leandrolopesm/eunuch/lib"

	"charm.land/bubbles/v2/textinput"
)

type readline struct {
    reader textinput.Model
    active bool
}

type textEngine struct {
	engine   lib.Engine
	passages map[string](Passage)

	currPassage string
	choiceIdx   int

	storyFile string

    lineReader readline
}

func (m *textEngine) loadStory() {
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

	m.passages["__unknown_passage"] = Passage {
		text: "ERROR: That choice linked to a non-existant passage\nSee the log file for more details",
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

func (m *textEngine) commitChoice() {
	currPassage := m.passages[m.currPassage]

    if len(currPassage.choices) != 0 {
        choice := currPassage.choices[m.choiceIdx]

        if hook,err := choice.hook.Try(); err == nil {
            if err := m.engine.Evaluate(core.MkScheme(hook)); err != nil {
                log.Errorf("Passage %s: Failed run hook for choice '%s': %v", m.currPassage, choice.value, err)
            }
        }

        last := m.currPassage
        if _, ok := m.passages[choice.target]; !ok {
            log.Errorf("Undefined passage %v", choice.target)
			m.currPassage = "__unknown_passage"
        } else {
            m.currPassage = choice.target
        }

        if hook,err := currPassage.hook.Try(); err == nil {
            if err := m.engine.Evaluate(core.MkScheme(hook)); err != nil {
                log.Errorf("Failed to load '%s': %v", choice.target, err)
                currPassage = m.passages[last]
            }
        }
    }
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

    if t.lineReader.active {
        str += t.lineReader.reader.View()
    }

    str += "\nPress 'q' to exit."

    return tryWrapLines(str)
}