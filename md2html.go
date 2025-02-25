package main

import (
	"html"
	"strings"
)

var (
	tagHTML = map[rune]string{
		'_': "i",
		'*': "b",
		'`': "code",
	}
	allMdChars = []rune{'_', '*', '`', '[', ']', '(', ')', '\\'}
)

const (
	btnURLPrefix   = "buttonurl:"
	sameLineSuffix = ":same"
)

var defaultConverter = Converter{
	BtnPrefix:      btnURLPrefix,
	SameLineSuffix: sameLineSuffix,
}

type Button struct {
	Name     string
	Content  string
	SameLine bool
}

type Converter struct {
	BtnPrefix      string
	SameLineSuffix string
}

func New() *Converter {
	return &Converter{
		BtnPrefix:      btnURLPrefix,
		SameLineSuffix: sameLineSuffix,
	}
}

func MD2HTML(input string) string {
	text, _ := defaultConverter.md2html([]rune(html.EscapeString(input)), false)
	return text
}

func (cv *Converter) MD2HTML(input string) string {
	text, _ := cv.md2html([]rune(html.EscapeString(input)), false)
	return text
}

func MD2HTMLButtons(input string) (string, []Button) {
	return defaultConverter.md2html([]rune(html.EscapeString(input)), true)
}

func (cv *Converter) MD2HTMLButtons(input string) (string, []Button) {
	return cv.md2html([]rune(html.EscapeString(input)), true)
}

func Reverse(s string, buttons []Button) string {
	return defaultConverter.Reverse(s, buttons)
}

func (cv *Converter) Reverse(s string, buttons []Button) string {
	return cv.reverse([]rune(s), buttons)
}

func StripMD(s string) string {
	return defaultConverter.StripMD(s)
}

func (cv *Converter) StripMD(s string) string {
	text, _ := cv.MD2HTMLButtons(s)
	return cv.stripHTML([]rune(text))
}

func StripHTML(s string) string {
	return defaultConverter.stripHTML([]rune(s))
}

func (cv *Converter) StripHTML(s string) string {
	return cv.stripHTML([]rune(s))
}

// todo: ``` support? -> add \n char to md chars and hence on \n, skip
func (cv *Converter) md2html(input []rune, buttons bool) (string, []Button) {
	var output strings.Builder
	v := map[rune][]int{}
	var containedMDChars []rune
	escaped := false
	offset := 0
	lastSync := 0
	var newInput []rune

	for pos, char := range input {
		if escaped {
			escaped = false
			continue
		}
		switch char {
		case '_', '*', '`', '[', ']', '(', ')':
			v[char] = append(v[char], pos-offset)
			containedMDChars = append(containedMDChars, char)
		case '\\':
			if len(input) <= pos+1 || !contains(input[pos+1], allMdChars) {
				continue
			}
			escaped = true
			newInput = append(newInput, input[lastSync:pos]...)
			offset++
			lastSync = pos + 1
		}
	}

	input = append(newInput, input[lastSync:]...)

	prev := 0
	var btnPairs []Button
	for i := 0; i < len(containedMDChars) && prev < len(input); i++ {
		currChar := containedMDChars[i] // check current character
		posArr := v[currChar]           // get the list of positions for relevant character
		if len(posArr) <= 0 {           // if no positions left, pass (sanity check)
			continue
		}
		// if we're past the currChar position, pass and update
		if posArr[0] < prev {
			v[currChar] = posArr[1:] // more sanity checks; if position is past, move on
			continue
		}
		switch currChar {
		case '_', '*', '`':
			// if fewer than 2 chars left, pass
			if len(posArr) < 2 {
				continue
			}

			bkp := map[rune][]int{}
			cnt := i // copy i to avoid changing if false
			// skip i to next same char (hence jumping all inbetween) (could be done with a normal range and continues?)
			// todo: OOB check on +1?
			for _, val := range containedMDChars[cnt+1:] {
				cnt++
				if val == currChar {
					break
				}

				if x, ok := bkp[val]; ok {
					bkp[val] = x[1:]
				} else {
					bkp[val] = v[val][1:]
				}
			}
			// pop currChar
			fstPos, rest := posArr[0], posArr[1:]
			v[currChar] = rest

			if !validStart(fstPos, input) {
				continue
			}
			ok := false
			var skipped int
			var sndPos int
			for idx, tmpSndPos := range rest {
				if validEnd(tmpSndPos, input) {
					rest = rest[idx+1:]
					sndPos = tmpSndPos
					ok = true
					break
				}
				skipped++
			}
			if !ok {
				continue
			}

			bkp[currChar] = rest
			output.WriteString(string(input[prev:fstPos]))
			output.WriteString("<" + tagHTML[currChar] + ">")
			output.WriteString(string(input[fstPos+1 : sndPos]))
			output.WriteString("</" + tagHTML[currChar] + ">")
			prev = sndPos + 1

			// ensure that items skipped for sndpos balance out with the count
			for _, m := range containedMDChars[cnt+1:] {
				if skipped <= 0 {
					break
				}
				if m == currChar {
					skipped--
				}
				cnt++
			}
			i = cnt // set i to copy

			for x, y := range bkp {
				v[x] = y
			}

		case '[':
			nameOpen, rest := posArr[0], posArr[1:]
			v['['] = rest
			if len(v[']']) < 1 || len(v['(']) < 1 || len(v[')']) < 1 || nameOpen < prev {
				continue
			}

			var nextNameClose int
			var nextLinkOpen int
			var nextLinkClose int

			wastedLinkClose := v[')']
		LinkLabel:
			for _, closeLinkPos := range v[')'] {
				if closeLinkPos > nameOpen {
					wastedLinkOpen := v['(']
					for _, openLinkpos := range v['('] {
						if openLinkpos > nameOpen && openLinkpos < closeLinkPos {
							wastedNameClose := v[']']
							for _, closeNamePos := range v[']'] {
								if closeNamePos == openLinkpos-1 {
									nextNameClose = closeNamePos
									nextLinkOpen = openLinkpos
									nextLinkClose = closeLinkPos
									v[']'] = wastedNameClose
									v['('] = wastedLinkOpen
									v[')'] = wastedLinkClose

									break LinkLabel
								}
								wastedNameClose = wastedNameClose[1:]
							}
						}
						wastedLinkOpen = wastedLinkOpen[1:]
					}
				}
				wastedLinkClose = wastedLinkClose[1:]
			}
			if nextLinkClose == 0 {
				// invalid
				continue
			}

			output.WriteString(string(input[prev:nameOpen]))
			link := string(input[nextLinkOpen+1 : nextLinkClose])
			name := string(input[nameOpen+1 : nextNameClose])
			if buttons && strings.HasPrefix(link, cv.BtnPrefix) {
				// is a button
				sameline := strings.HasSuffix(link, cv.SameLineSuffix)
				if sameline {
					link = link[:len(link)-len(cv.SameLineSuffix)]
				}
				btnPairs = append(btnPairs, Button{
					Name:     html.UnescapeString(name),
					Content:  strings.TrimLeft(strings.TrimSpace(link[len(cv.BtnPrefix):]), "/"),
					SameLine: sameline,
				})
			} else {
				output.WriteString(`<a href="` + link + `">` + name + `</a>`)
			}

			prev = nextLinkClose + 1
		}
	}
	output.WriteString(string(input[prev:]))

	return strings.TrimSpace(output.String()), btnPairs
}

// NOTE: If this gets edited, make sure to edit the strip() method too
// TODO: this needs to return string, error to handle bad parsing
func (cv *Converter) reverse(r []rune, buttons []Button) string {
	prev := 0
	out := strings.Builder{}
	for i := 0; i < len(r); i++ {
		switch r[i] {
		case '<':
			closeTag := 0
			for ix, c := range r[i+1:] {
				if c == '>' {
					closeTag = i + ix + 1
					break
				}
			}
			if closeTag == 0 {
				// "no close tag"
				return ""
			}

			closingOpen := 0
			for ix, c := range r[closeTag:] {
				if c == '<' {
					closingOpen = closeTag + ix
					break
				}
			}
			if closingOpen == 0 {
				// "no closing open"
				return ""
			}

			closingClose := 0
			for ix, c := range r[closingOpen:] {
				if c == '>' {
					closingClose = closingOpen + ix
					break
				}
			}
			if closingClose == 0 {
				// "no closing close"
				return ""
			}

			tag := string(r[i+1 : closeTag])
			// todo: check expected closing tag
			out.WriteString(html.UnescapeString(string(r[prev:i])))
			if link.MatchString(tag) {
				matches := link.FindStringSubmatch(tag)
				out.WriteString("[" + html.UnescapeString(EscapeMarkdown(r[closeTag+1:closingOpen], []rune{'[', ']', '(', ')'})) + "](" + matches[1] + ")")
			} else if tag == "b" {
				out.WriteString("*" + html.UnescapeString(EscapeMarkdown(r[closeTag+1:closingOpen], []rune{'*'})) + "*")
			} else if tag == "i" {
				out.WriteString("_" + html.UnescapeString(EscapeMarkdown(r[closeTag+1:closingOpen], []rune{'_'})) + "_")
			} else if tag == "code" {
				out.WriteString("`" + html.UnescapeString(EscapeMarkdown(r[closeTag+1:closingOpen], []rune{'`'})) + "`")
			} else {
				// unknown tag
				return ""
			}
			prev = closingClose + 1
			i = closingClose

		case '\\', '_', '*', '`', '[', ']', '(', ')': // these all need to be escaped to ensure we retain the same message
			out.WriteString(html.UnescapeString(string(r[prev:i])))
			out.WriteRune('\\')
			out.WriteRune(r[i])
			prev = i + 1
		}
	}
	out.WriteString(html.UnescapeString(string(r[prev:])))

	for _, btn := range buttons {
		out.WriteString("\n[" + btn.Name + "](buttonurl://" + html.UnescapeString(btn.Content))
		if btn.SameLine {
			out.WriteString(":same")
		}
		out.WriteString(")")
	}

	return out.String()
}

func (cv *Converter) stripHTML(r []rune) string {
	prev := 0
	out := strings.Builder{}
	for i := 0; i < len(r); i++ {
		switch r[i] {
		case '<':
			closeTag := 0
			for ix, c := range r[i+1:] {
				if c == '>' {
					closeTag = i + ix + 1
					break
				}
			}
			if closeTag == 0 {
				// "no close tag"
				return ""
			}

			closingOpen := 0
			for ix, c := range r[closeTag:] {
				if c == '<' {
					closingOpen = closeTag + ix
					break
				}
			}
			if closingOpen == 0 {
				// "no closing open"
				return ""
			}

			closingClose := 0
			for ix, c := range r[closingOpen:] {
				if c == '>' {
					closingClose = closingOpen + ix
					break
				}
			}
			if closingClose == 0 {
				// "no closing close"
				return ""
			}

			out.WriteString(html.UnescapeString(string(r[prev:i])))
			out.WriteString(html.UnescapeString(string(r[closeTag+1 : closingOpen])))

			prev = closingClose + 1
			i = closingClose
		}
	}
	out.WriteString(html.UnescapeString(string(r[prev:])))

	return out.String()
}

func EscapeMarkdown(r []rune, toEscape []rune) string {
	out := strings.Builder{}
	for i, x := range r {
		if contains(x, toEscape) {
			if i == 0 || i == len(r)-1 || validEnd(i, r) || validStart(i, r) {
				out.WriteRune('\\')
			}
		}
		out.WriteRune(x)
	}
	return out.String()
}
