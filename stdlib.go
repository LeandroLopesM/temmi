package main

import (
	"fmt"
	"slices"
	"strings"

	"github.com/leandrolopesm/eunuch/core"
	"github.com/leandrolopesm/eunuch/engine"
	"github.com/leandrolopesm/eunuch/util"
	"github.com/logrusorgru/aurora/v4"
)

var ActivePassage Passage
var Model *textEngine

func Register(eng *textEngine) {
	Model = eng

	eng.engine.AddFunc(
        "set-opt!",
        engine.Builtin{
            Args: []core.Type{core.Symbol, core.Any},
            Call: SetOpt,
            Prepare: util.Some[engine.StagingFunc](
                func(s core.Scheme, e *engine.Engine) error {
		            e.Push(s.Params()[0])
		            e.Evaluate(s.Params()[1])

                    return nil
	            },
            ),
        },
    )
	eng.engine.AddFunc(
        "passage",
        engine.Builtin{
            VarArgs: true,
            Call: NewPassage,
            Prepare: util.Some[engine.StagingFunc](
                func(s core.Scheme, e *engine.Engine) error {
                    for _,arg := range s.Params() {
                        e.Push(arg)
                    }
                    return nil
                },
            ),
        },
    )
	eng.engine.AddFunc(
		"on-load",
		engine.Builtin{
			VarArgs: true,
			Call: LoadHook,
			Prepare: util.Some[engine.StagingFunc](
				func(s core.Scheme, e *engine.Engine) error {
					for _,arg := range s.Params() {
						e.Push(arg)
					}

					return nil
				},
			),
		},
	)

	eng.engine.AddFunc("text", engine.Builtin{VarArgs: true, Call: NewText})
		eng.engine.AddFunc("ul", engine.Builtin{VarArgs: true, Call: UnderlineText})
		eng.engine.AddFunc("bold", engine.Builtin{VarArgs: true, Call: BoldText})
	eng.engine.AddFunc("choice", engine.Builtin{VarArgs: true, Call: NewChoice})
	eng.engine.AddFunc("readline", engine.Builtin{Call: Readline})

	eng.engine.AddFunc("dark-gray", engine.Builtin{Args: []core.Type{core.String}, Call: ColorFn("dark-gray")})
	eng.engine.AddFunc("green", engine.Builtin{Args: []core.Type{core.String}, Call: ColorFn("green")})
	eng.engine.AddFunc("red", engine.Builtin{Args: []core.Type{core.String}, Call: ColorFn("red")})
}

func ColorFn(color string) engine.BuiltinExec {
	return func(e *engine.Engine) error {
		str,_ := e.Pop()
		switch color {
		case "red":
			e.Push(core.MkString(aurora.Red(str.Value.(string)).String()))
		case "dark-gray":
			e.Push(core.MkString(aurora.BrightBlack(str.Value.(string)).String()))
		case "green":
			e.Push(core.MkString(aurora.Green(str.Value.(string)).String()))
		default:
			return fmt.Errorf("Unknown color '%s'", color)
		}

		return nil
	}
}

func Readline(eng *engine.Engine) error {
	Model.lineReader.active = true
	Model.lineReader.reader.Focus()

	return nil
}

func SetOpt(eng *engine.Engine) error {
	val, _ := eng.Pop()
	name, _ := eng.Pop()

	switch name.Value.(string) {
	case "responsive":
		optResponsive = val.Value.(bool)
	case "wrapLines":
		optWrapLines = val.Value.(bool)
	case "lineWidth":
		optLineWidth = int(val.Value.(int64))
	default:
		return fmt.Errorf("Unknown option '%s'", name.Value.(string))
	}

	return nil
}

func LoadHook(eng *engine.Engine) error {
	ActivePassage.hook = util.Some(util.Assert(eng.Pop()).Value.(core.Scheme))
	return nil
}

func NewChoice(eng *engine.Engine) error {
	first := util.Assert(eng.Pop())
	var hook core.Scheme
	var text string
	if first.Type == core.SchemeType {
		hook = first.Value.(core.Scheme)
		text,_ = util.Assert(eng.Pop()).Value.(string)
	} else {
		text,_ = first.Value.(string)
	}
	target := util.Assert(eng.Pop()).Value.(string)

	if val, err := eng.Pop(); err == nil {
		hook = val.Value.(core.Scheme)
	}

	ActivePassage.choices = append(ActivePassage.choices, Choice{
		value:  text,
		target: target,
		hook:   &hook,
	})

	return nil
}

func collectText(eng *engine.Engine) string {
	args := varArgAsArgv(eng)
	var out string
	for _,arg := range args {
		out += core.SprintUnit(arg) + " "
	}
	// Remove trailing ' '
	out = string([]rune(out)[:len(out)-1])
	return out
}

func UnderlineText(eng *engine.Engine) error {
	var out string = collectText(eng)
	eng.Push(core.MkString(aurora.Underline(out).String()))
	return nil
}

func BoldText(eng *engine.Engine) error {
	var out string = collectText(eng)
	eng.Push(core.MkString(aurora.Bold(out).String()))
	return nil
}

func NewText(eng *engine.Engine) error {
	var args []core.Unit

	for {
		if val, err := eng.Pop(); err != nil {
			break
		} else {
			args = append(args, val)
		}
	}

	idx := len(args) - 1

	var stringOut string
	for idx >= 0 {
		stringOut += core.SprintUnit(args[idx])
		idx--
	}

    stringOut = strings.ReplaceAll(stringOut, "\\n", "\n")
	ActivePassage.text += stringOut

	return nil
}

func NewPassage(eng *engine.Engine) error {
	ActivePassage = Passage{}
	args := varArgAsArgv(eng)
	passageName := args[0].Value.(string)

	for _, arg := range args[0:] {
		if err := eng.Evaluate(arg); err != nil {
			return err
		}
	}

	Model.passages[passageName] = ActivePassage
	ActivePassage = Passage{}

	return nil
}

func varArgAsArgv(eng *engine.Engine) []core.Unit {
	var args []core.Unit

	for {
		if val, err := eng.Pop(); err != nil {
			break
		} else {
			args = append(args, val)
		}
	}

	slices.Reverse(args)

	return args
}
