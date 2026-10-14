package main

import (
	"strings"
)

func autoWrapLines(rawStr string, maxWidth int) string {
	if len(rawStr) < maxWidth {
		return rawStr
	}

	var out string

	for _, line := range strings.Split(rawStr, "\n") {
		var brokenLine string

		spaceLeft := maxWidth

		if len(line) > maxWidth {
			for _, word := range strings.Split(line, " ") {
				if len(word)+1 > spaceLeft {
					brokenLine += "\n"
					spaceLeft = maxWidth - len(word)
				} else {
					brokenLine += " "
					spaceLeft -= len(word) + 1
				}

                brokenLine += word
			}
		} else {
            brokenLine = line
        }

		out += "\n" + brokenLine
	}

	return out
}

func tryWrapLines(str string) string {
    if optWrapLines {
		if optResponsive {
			return autoWrapLines(str, windowSize.Width)
		} else {
			return autoWrapLines(str, optLineWidth)
		}
	}

    return str
}