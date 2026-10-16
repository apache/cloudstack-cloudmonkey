package readline

import (
	"testing"
)

type mapCompletionTestCompleter struct{}

func (mapCompletionTestCompleter) Do(_ []rune, _ int) ([][]rune, int) {
	return [][]rune{
		[]rune("[0].key="),
		[]rune("[0].value="),
	}, 0
}

func TestMapCompletionReplacesTrailingEqual(t *testing.T) {
	initial := "list virtualmachines tags="

	cfg := &Config{
		AutoComplete: mapCompletionTestCompleter{},
	}

	term := &terminal{}
	term.cfg.Store(cfg)
	term.dimensions.Store(&termDimensions{
		width:  80,
		height: 24,
	})

	op := &operation{
		t: term,
		buf: &runeBuffer{
			buf: []rune(initial),
			idx: len([]rune(initial)),
			w:   term,
		},
	}
	op.completer = newOpCompleter(term, op)

	op.completer.OnComplete()

	got := string(op.buf.Runes())
	want := "list virtualmachines tags[0]."

	if got != want {
		t.Fatalf("after Tab, expected %q, got %q", want, got)
	}
}
