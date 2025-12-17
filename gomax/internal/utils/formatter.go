package utils

import (
	"regexp"
	"strings"

	"github.com/fresh-milkshake/gomax/enums"
	"github.com/fresh-milkshake/gomax/types"
)

var markupBlockPattern = regexp.MustCompile(
	`\*\*(?P<strong>.+?)\*\*|\*(?P<italic>.+?)\*|__(?P<underline>.+?)__|~~(?P<strike>.+?)~~`,
)

// Парсит markdown и возвращает элементы форматирования и очищенный текст.
func GetElementsFromMarkdown(text string) ([]types.Element, string) {
	text = strings.Trim(text, "\n")
	var elements []types.Element
	var cleanParts []string
	currentPos := 0
	lastEnd := 0

	matches := markupBlockPattern.FindAllStringSubmatchIndex(text, -1)
	for _, match := range matches {
		between := text[lastEnd:match[0]]
		if between != "" {
			cleanParts = append(cleanParts, between)
			currentPos += len(between)
		}

		var innerText string
		var fmtType enums.FormattingType

		strongIdx := markupBlockPattern.SubexpIndex("strong")
		italicIdx := markupBlockPattern.SubexpIndex("italic")
		underlineIdx := markupBlockPattern.SubexpIndex("underline")
		strikeIdx := markupBlockPattern.SubexpIndex("strike")

		if strongIdx >= 0 && match[strongIdx*2] >= 0 {
			innerText = text[match[strongIdx*2]:match[strongIdx*2+1]]
			fmtType = enums.FormattingTypeStrong
		} else if italicIdx >= 0 && match[italicIdx*2] >= 0 {
			innerText = text[match[italicIdx*2]:match[italicIdx*2+1]]
			fmtType = enums.FormattingTypeEmphasized
		} else if underlineIdx >= 0 && match[underlineIdx*2] >= 0 {
			innerText = text[match[underlineIdx*2]:match[underlineIdx*2+1]]
			fmtType = enums.FormattingTypeUnderline
		} else if strikeIdx >= 0 && match[strikeIdx*2] >= 0 {
			innerText = text[match[strikeIdx*2]:match[strikeIdx*2+1]]
			fmtType = enums.FormattingTypeStrikethrough
		}

		if innerText != "" {
			nextPos := match[1]
			hasNewline := (nextPos < len(text) && text[nextPos] == '\n') || nextPos == len(text)

			length := len(innerText)
			if hasNewline {
				length++
			}

			from := currentPos
			elements = append(elements, types.Element{
				Type:   fmtType,
				From:   &from,
				Length: length,
			})

			cleanParts = append(cleanParts, innerText)
			if hasNewline {
				cleanParts = append(cleanParts, "\n")
			}

			currentPos += length

			if nextPos < len(text) && text[nextPos] == '\n' {
				lastEnd = match[1] + 1
			} else {
				lastEnd = match[1]
			}
		} else {
			lastEnd = match[1]
		}
	}

	tail := text[lastEnd:]
	if tail != "" {
		cleanParts = append(cleanParts, tail)
	}

	cleanText := strings.Join(cleanParts, "")
	return elements, cleanText
}
