package text

import (
	"strings"

	"github.com/hajimehoshi/ebiten/v2/text/v2"
)

func CurrentIndexInTextReveal(startTick, currentTick, ticksPerCharacter int, text string) (finished bool, count int) {
	if currentTick < startTick {
		return false, 0
	}

	if ticksPerCharacter <= 0 {
		return true, len(text)
	}

	elapsedTicks := currentTick - startTick
	count = (elapsedTicks / ticksPerCharacter) + 1

	if count >= len(text) {
		return true, len(text)
	}

	return false, count
}

func WrapText(s string, face *text.GoTextFace, maxWidth float64) string {
	var result strings.Builder
	var currentLine strings.Builder

	words := strings.Fields(s)
	if len(words) == 0 {
		return ""
	}

	currentLine.WriteString(words[0])

	for i := 1; i < len(words); i++ {
		word := words[i]
		potentialLine := currentLine.String() + " " + word

		advance, _ := text.Measure(potentialLine, face, face.Metrics().HAscent)

		if advance > maxWidth {
			result.WriteString(currentLine.String())
			result.WriteString("\n")
			currentLine.Reset()
			currentLine.WriteString(word)
		} else {
			currentLine.WriteString(" ")
			currentLine.WriteString(word)
		}
	}

	result.WriteString(currentLine.String())
	return result.String()
}

func ShouldPlayRevealSound(startTick, currentTick, ticksPerCharacter int, text string) bool {
	if currentTick < startTick || ticksPerCharacter <= 0 {
		return false
	}

	elapsedTicksNow := currentTick - startTick
	countNow := (elapsedTicksNow / ticksPerCharacter)

	elapsedTicksPrev := (currentTick - 1) - startTick
	if elapsedTicksPrev < 0 {
		return true
	}
	countPrev := (elapsedTicksPrev / ticksPerCharacter)

	return countNow > countPrev && countNow < len(text)
}
