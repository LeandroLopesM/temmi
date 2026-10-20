package main

import (
	"math"
	"slices"
	"strings"
)

// ==================================================================================================
// ADAPTED FROM: https://every-algorithm.github.io/2024/10/03/knuthplass_line-breaking_algorithm.html
// ==================================================================================================

// Knuth–Plass line-breaking algorithm
// This implementation finds the optimal set of line breaks for a paragraph
// given a list of words and a maximum line width. It uses dynamic programming
// to compute the minimal total badness, where badness is defined as the cube
// of the extra space left on a line.

func KnuthPlass(words []string, max_width int) []string {
    n := len(words)
    // Compute the width of each word and the cumulative widths
    var wordWidths []int
    for _,w := range words {
        wordWidths = append(wordWidths, len(w))
    }

    cum_widths := make([]int, n + 1)

    for i := range n {
        cum_widths[i + 1] = cum_widths[i] + wordWidths[i] + 1  // +1 for space
    }

    INF := 999999999
    // best[i] = minimal total badness for words[0:i]
    best := make([]int, n + 1)
    for i := range best {
        best[i] = INF
    }

    best[0] = 0

    // prev[i] = index of the previous line break before word i
    prev := make([]int, n + 1)
    for i := range prev {
        prev[i] = -1
    }

    for i := 1; i < n + 1; i++ {
        for j := range i {
            // Total width of words[j:i] including spaces
            line_len := cum_widths[i] - cum_widths[j] - 1

            if line_len > max_width {
                continue
            }

            var badness int
            if i == n {
                badness = 0  // No badness for the last line
            } else {
                extra := max_width - line_len
                badness = int(math.Pow(float64(extra), 3))
            }

            cost := best[j] + badness
            if cost < best[i] {
                best[i] = cost
                prev[i] = j
            }
        }
    }

    // Reconstruct the lines
    var lines []string
    idx := n
    for idx > 0 {
        j := prev[idx]
        lines = append(lines, strings.Join(words[j:idx], " "))
        idx = j
    }

    slices.Reverse(lines)
    return lines
}