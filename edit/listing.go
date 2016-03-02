package edit

import "unicode/utf8"

// listing implements a listing mode that supports the notion of selecting an
// entry and filtering entries.
type listing struct {
	typ      ModeType
	provider listingProvider
	selected int
	filter   string
}

type listingProvider interface {
	Len() int
	Show(i, w int) string
	Filter(filter string) int
	Accept(i int, ed *Editor)
	ModeTitle(int) string
}

func (l *listing) Mode() ModeType {
	return l.typ
}

func (l *listing) ModeLine(width int) *buffer {
	title := l.provider.ModeTitle(l.selected)
	// TODO keep it one line.
	b := newBuffer(width)
	b.writes(TrimWcWidth(title, width), styleForMode)
	b.writes(" ", "")
	b.writes(l.filter, styleForFilter)
	b.dot = b.cursor()
	return b
}

func (l *listing) List(width, maxHeight int) *buffer {
	n := l.provider.Len()
	b := newBuffer(width)
	if n == 0 {
		b.writes(TrimWcWidth("(no result)", width), "")
		return b
	}
	low, high := findWindow(n, l.selected, maxHeight)
	for i := low; i < high; i++ {
		if i > low {
			b.newline()
		}
		style := ""
		if i == l.selected {
			style = styleForSelected
		}
		b.writes(TrimWcWidth(l.provider.Show(i, width), width), style)
	}
	return b
}

func (l *listing) changeFilter(newfilter string) {
	l.filter = newfilter
	l.selected = l.provider.Filter(newfilter)
}

func (l *listing) backspace() bool {
	_, size := utf8.DecodeLastRuneInString(l.filter)
	if size > 0 {
		l.changeFilter(l.filter[:len(l.filter)-size])
		return true
	}
	return false
}

func (l *listing) prev(cycle bool) {
	n := l.provider.Len()
	if n == 0 {
		return
	}
	l.selected--
	if l.selected == -1 {
		if cycle {
			l.selected += n
		} else {
			l.selected++
		}
	}
}

func (l *listing) next(cycle bool) {
	n := l.provider.Len()
	if n == 0 {
		return
	}
	l.selected++
	if l.selected == n {
		if cycle {
			l.selected -= n
		} else {
			l.selected--
		}
	}
}

func (l *listing) accept(ed *Editor) {
	if l.selected >= 0 {
		l.provider.Accept(l.selected, ed)
	}
}

func (l *listing) handleFilterKey(k Key) bool {
	if likeChar(k) {
		l.changeFilter(l.filter + string(k.Rune))
		return true
	}
	return false
}

func (l *listing) defaultBinding(ed *Editor) {
	if !l.handleFilterKey(ed.lastKey) {
		startInsert(ed)
		ed.nextAction = action{typ: reprocessKey}
	}
}

func addListingBuiltins(prefix string, l func(*Editor) *listing) {
	add := func(name string, f func(*Editor)) {
		builtins = append(builtins, Builtin{prefix + name, f})
	}
	add("prev", func(ed *Editor) { l(ed).prev(false) })
	add("prev-cycle", func(ed *Editor) { l(ed).prev(true) })
	add("next", func(ed *Editor) { l(ed).next(false) })
	add("next-cycle", func(ed *Editor) { l(ed).next(true) })
	add("backspace", func(ed *Editor) { l(ed).backspace() })
	add("accept", func(ed *Editor) { l(ed).accept(ed) })
	add("default", func(ed *Editor) { l(ed).defaultBinding(ed) })
}

func addListingDefaultBindings(prefix string, m ModeType) {
	add := func(k Key, name string) {
		defaultBindings[m][k] = prefix + name
	}
	add(Key{Up, 0}, "prev")
	add(Key{Down, 0}, "next")
	add(Key{Tab, 0}, "next-cycle")
	add(Key{Backspace, 0}, "backspace")
	add(Key{Enter, 0}, "accept")
	add(Default, "default")
	defaultBindings[m][Key{'[', Ctrl}] = "start-insert"
}
