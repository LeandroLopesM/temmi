package main

import (
	"errors"
	"fmt"
	"strings"

	"github.com/leandrolopesm/eunuch/core"
	"github.com/leandrolopesm/eunuch/engine"
	"github.com/leandrolopesm/eunuch/util"
	"github.com/logrusorgru/aurora/v4"
)

var Model *textEngine
var PassageBuilderTarget *Passage

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
            Call: SetOpt,
            Argc: 2,
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
            Call: NewPassage,
            VarArgs: true,
            Prepare: dontEval,
        },
    )
	eng.engine.AddFunc(
		"on-load",
		engine.Builtin{
			Call: LoadHook,
			VarArgs: true,
			Prepare: dontEval,
		},
	)
	eng.engine.AddFunc("set!", engine.Builtin{Argc: 2, Call: Set, Prepare: dontEval})

	eng.engine.AddFunc("set-text!", engine.Builtin{VarArgs: true, Call: SetText})
	eng.engine.AddFunc("if", engine.Builtin{VarArgs: true, Call: If, Prepare: dontEval})

	eng.engine.AddFunc("text", engine.Builtin{VarArgs: true, Call: NewText})
		eng.engine.AddFunc("ul", engine.Builtin{VarArgs: true, Call: UnderlineText})
		eng.engine.AddFunc("bold", engine.Builtin{VarArgs: true, Call: BoldText})
	eng.engine.AddFunc("choice", engine.Builtin{VarArgs: true, Call: NewChoice})

	eng.engine.AddFunc("begin", engine.Builtin{VarArgs: true, Call: Begin})

	eng.engine.AddFunc("clear-text!", engine.Builtin{Argc: 0, Call: ClearText})

	eng.engine.AddFunc("dark-gray", engine.Builtin{Argc: 1, Call: ColorFn("dark-gray")})
	eng.engine.AddFunc("green", engine.Builtin{Argc: 1, Call: ColorFn("green")})
	eng.engine.AddFunc("red", engine.Builtin{Argc: 1, Call: ColorFn("red")})
}

func Set(e *engine.Engine) error {
	val,_ := e.Pop()
	if err := e.Evaluate(val); err != nil {
		return err
	}

	newVal := util.Assert(e.Pop())
	name := util.Assert(e.Pop()).(core.Symbol)
	e.SetVar(string(name), newVal)
	return nil
}

func Begin(e *engine.Engine) error {
	return nil
}

func SetText(e *engine.Engine) error  {
	args := e.GenArgv()

	var out = strings.Builder{}
	for _,arg := range args {
		out.WriteString(arg.ToString())
	}

	Model.currPassage.text = out.String();

	return nil
}

func ClearText(e *engine.Engine) error  {
	currPassage := Model.currPassage;
	currPassage.text = "";
	Model.currPassage = currPassage

	return nil
}

func If(e *engine.Engine) error  {
	args := e.GenArgv()
	e.Evaluate(args[0])

	var cond bool

	if v,err := e.Pop(); err != nil || v.TypeString() != "Bool" {
		return errors.New("condition did not yield a boolean value")
	} else {
		cond = bool(v.(core.Bool))
	}

	onTrue := args[1].(core.Scheme)

	var onFalse core.Scheme
	if len(args) == 3 {
		onFalse = args[2].(core.Scheme)
	}

	if cond {
		e.Evaluate(core.Scheme(onTrue))
	} else if len(onFalse.Args) != 0 {
		e.Evaluate(core.Scheme(onFalse))
	}

	return nil
}

func ColorFn(color string) engine.BuiltinExec {
	return func(e *engine.Engine) error {
		str,_ := e.Pop()
		switch color {
		case "red":
			e.Push(core.String(aurora.Red(str.ToString()).String()))
		case "dark-gray":
			e.Push(core.String(aurora.BrightBlack(str.ToString()).String()))
		case "green":
			e.Push(core.String(aurora.Green(str.ToString()).String()))
		default:
			return fmt.Errorf("unknown color '%s'", color)
		}

		return nil
	}
}

func SetOpt(eng *engine.Engine) error {
	val, _ := eng.Pop()
	name, _ := eng.Pop()

	switch name.(core.Symbol) {
	case "responsive":
		optResponsive = bool(val.(core.Bool))
	case "wrapLines":
		optWrapLines = bool(val.(core.Bool))
	case "lineWidth":
		optLineWidth = int(val.(core.Integer))
	default:
		return fmt.Errorf("Unknown option '%s'", name.ToString())
	}

	return nil
}

func LoadHook(eng *engine.Engine) error {
	Model.currPassage.hook = util.Assert(eng.Pop()).(core.Scheme)
	return nil
}

func NewChoice(eng *engine.Engine) error {
	args := eng.GenArgv()
	target := args[0].(core.String)
	text := args[1].(core.String)

	var hook core.Value
	if len(args) > 2 {
		hook = args[2]
	} else {
		hook = nil
	}

	Model.currPassage.choices = append(Model.currPassage.choices, Choice{
		value:  string(text),
		target: string(target),
		hook:   hook,
	})

	return nil
}

func collectText(eng *engine.Engine) string {
	args := eng.GenArgv()
	var out = strings.Builder{}
	for _,arg := range args {
		// The linter SWEARS this is more efficient
		// than out += core.(...) + " "
		out.WriteString(arg.ToString());
		out.WriteString(" ")
	}
	// Remove trailing ' '
	outStr := out.String()
	return string([]rune(outStr)[:len(outStr)-1])
}

func UnderlineText(eng *engine.Engine) error {
	var out = collectText(eng)
	eng.Push(core.String(aurora.Underline(out).String()))
	return nil
}

func BoldText(eng *engine.Engine) error {
	var out = collectText(eng)
	eng.Push(core.String(aurora.Bold(out).String()))
	return nil
}

func NewText(eng *engine.Engine) error {
	var args []core.Value

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
		stringOut.WriteString(args[idx].ToString())
		idx--
	}

	Model.currPassage.text += strings.ReplaceAll(stringOut.String(), "\\n", "\n")

	return nil
}

func NewPassage(eng *engine.Engine) error {
	Model.currPassage = Passage{}
	args := eng.GenArgv()
	passageName := args[0].(core.String)

	Model.passages[string(passageName)] = args[0:]

	return nil
}