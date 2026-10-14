package main

import (
	"errors"
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

var dontEval = util.Some[engine.StagingFunc](func (s core.Scheme, e *engine.Engine) error {
	for _,arg := range s.Params() {
		e.Push(arg)
	}
	return nil
})

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
		            return e.Evaluate(s.Params()[1])
	            },
            ),
        },
    )
	eng.engine.AddFunc(
        "passage",
        engine.Builtin{
            VarArgs: true,
            Call: NewPassage,
            Prepare: dontEval,
        },
    )
	eng.engine.AddFunc(
		"on-load",
		engine.Builtin{
			VarArgs: true,
			Call: LoadHook,
			Prepare: dontEval,
		},
	)
	eng.engine.AddFunc("set!", engine.Builtin{Args: []core.Type{ core.Symbol, core.Any }, Call: Set, Prepare: dontEval})

	eng.engine.AddFunc("set-text!", engine.Builtin{VarArgs: true, Call: SetText})
	eng.engine.AddFunc("if", engine.Builtin{VarArgs: true, Call: If, Prepare: dontEval})

	eng.engine.AddFunc("text", engine.Builtin{VarArgs: true, Call: NewText})
		eng.engine.AddFunc("ul", engine.Builtin{VarArgs: true, Call: UnderlineText})
		eng.engine.AddFunc("bold", engine.Builtin{VarArgs: true, Call: BoldText})
	eng.engine.AddFunc("choice", engine.Builtin{VarArgs: true, Call: NewChoice})
	eng.engine.AddFunc("readline", engine.Builtin{Call: Readline})

	eng.engine.AddFunc("begin", engine.Builtin{VarArgs: true, Call: Begin})

	eng.engine.AddFunc("clear-text!", engine.Builtin{Args: []core.Type{}, Call: ClearText})

	eng.engine.AddFunc("dark-gray", engine.Builtin{Args: []core.Type{core.String}, Call: ColorFn("dark-gray")})
	eng.engine.AddFunc("green", engine.Builtin{Args: []core.Type{core.String}, Call: ColorFn("green")})
	eng.engine.AddFunc("red", engine.Builtin{Args: []core.Type{core.String}, Call: ColorFn("red")})
}

func Set(e *engine.Engine) error {
	val,_ := e.Pop()
	if err := e.Evaluate(val); err != nil {
		return err
	}

	newVal := util.Assert(e.Pop())
	name := util.Assert(e.Pop()).Value.(string)
	e.SetVar(name, newVal)
	return nil
}

func Begin(e *engine.Engine) error {
	return nil
}

func SetText(e *engine.Engine) error  {
	args := varArgAsArgv(e)

	var out = strings.Builder{}
	for _,arg := range args {
		out.WriteString(core.SprintUnit(arg))
	}

	currPassage := Model.passages[Model.currPassage];
	currPassage.text = out.String();
	Model.passages[Model.currPassage] = currPassage

	return nil
}

func ClearText(e *engine.Engine) error  {
	currPassage := Model.passages[Model.currPassage];
	currPassage.text = "";
	Model.passages[Model.currPassage] = currPassage

	return nil
}

func If(e *engine.Engine) error  {
	args := varArgAsArgv(e)
	e.Evaluate(args[0])

	var cond bool

	if v,err := e.Pop(); err != nil || v.Type != core.Bool {
		return errors.New("condition did not yield a boolean value")
	} else {
		cond = v.Value.(bool)
	}

	onTrue := args[1].Value.(core.Scheme)

	var onFalse core.Scheme
	if len(args) == 3 {
		onFalse = args[2].Value.(core.Scheme)
	}

	if cond {
		e.Evaluate(core.MkScheme(onTrue))
	} else if len(onFalse.Args) != 0 {
		e.Evaluate(core.MkScheme(onFalse))
	}

	return nil
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
			return fmt.Errorf("unknown color '%s'", color)
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
	args := varArgAsArgv(eng)
	target := args[0].Value.(string)
	text := args[1].Value.(string)
	var hook util.Option[core.Scheme]
	if len(args) > 2 {
		hook = util.Some(args[2].Value.(core.Scheme))
	} else {
		hook = util.None[core.Scheme]()
	}

	ActivePassage.choices = append(ActivePassage.choices, Choice{
		value:  text,
		target: target,
		hook:   hook,
	})

	return nil
}

func collectText(eng *engine.Engine) string {
	args := varArgAsArgv(eng)
	var out = strings.Builder{}
	for _,arg := range args {
		// The linter SWEARS this is more efficient
		// than out += core.(...) + " "
		out.WriteString(core.SprintUnit(arg));
		out.WriteString(" ")
	}
	// Remove trailing ' '
	outStr := out.String()
	return string([]rune(outStr)[:len(outStr)-1])
}

func UnderlineText(eng *engine.Engine) error {
	var out = collectText(eng)
	eng.Push(core.MkString(aurora.Underline(out).String()))
	return nil
}

func BoldText(eng *engine.Engine) error {
	var out = collectText(eng)
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

	var stringOut = strings.Builder{}
	for idx >= 0 {
		stringOut.WriteString(core.SprintUnit(args[idx]))
		idx--
	}

	ActivePassage.text += strings.ReplaceAll(stringOut.String(), "\\n", "\n")

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
	ActivePassage = Passage{
		hook: util.None[core.Scheme](),
	}

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
