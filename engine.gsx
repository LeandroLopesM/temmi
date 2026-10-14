package main

import (
    "fmt"
    tui "github.com/grindlemire/go-tui"
)

type counter struct {
    count *tui.State[int]
}

func Counter() *counter {
    return &counter{count: tui.NewState(0)}
}

func (c *counter) KeyMap() tui.KeyMap {
    return tui.KeyMap{
        tui.On(tui.Rune('+'), func(ke tui.KeyEvent) { c.count.Update(func(v int) int { return v + 1 }) }),
        tui.On(tui.Rune('-'), func(ke tui.KeyEvent) { c.count.Update(func(v int) int { return v - 1 }) }),
        tui.On(tui.Rune('q'), func(ke tui.KeyEvent) { ke.App().Stop() }),
    }
}

templ (c *counter) Render() {
    <div class="flex-col items-center justify-center h-full gap-1">
        <span class="font-bold text-cyan">{fmt.Sprintf("Count: %d", c.count.Get())}</span>
        <span class="font-dim">+/- to change, q to quit</span>
    </div>
}