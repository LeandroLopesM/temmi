package main

import (
	"strings"
)

func GreedyWrap(rawStr string, maxWidth int) string {
	if len(rawStr) < maxWidth {
		return rawStr
	}

	var out = strings.Builder{}

	for line := range strings.SplitSeq(rawStr, "\n") {
		var brokenLine = strings.Builder{}

		spaceLeft := maxWidth

		if len(line) > maxWidth {
			for word := range strings.SplitSeq(line, " ") {
				if len(word)+1 > spaceLeft {
					brokenLine.WriteRune('\n')
					spaceLeft = maxWidth - len(word)
				} else {
					brokenLine.WriteRune(' ')
					spaceLeft -= len(word) + 1
				}

				brokenLine.WriteString(word)
			}
		} else {
			brokenLine.WriteString(line)
		}

		out.WriteRune('\n')
		out.WriteString(brokenLine.String())
	}

	return out.String()
}

func tryWrapLines(str string) string {
	if optWrapLines {
		var out string
		lines := strings.SplitSeq(str, "\n")

		for line := range lines {
			if strings.HasPrefix(line, "<img>") {
				out += line[5:] + "\n"
				continue
			}

			out += strings.Join(
				KnuthPlass(
					strings.Split(line, " "),
					optLineWidth),
				"\n")

			out += "\n"
		}

		return out
	}

	return str
}
