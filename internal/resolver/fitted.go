package resolver

import (
	"strconv"

	"github.com/rivo/uniseg"

	"github.com/kryptamine/herdr-auto-title/internal/state"
)

// positionMark separates a tab's position from the name that follows it. It is
// deliberately not Separator: the position is not one of the things a title
// says about a tab, and the parts separator would read as if it were.
const positionMark = " · "

// Fitted cuts what it wraps to the tab bar and, when asked, puts the tab's
// position in front — the key that switches to that tab. It wraps a resolver
// rather than being part of one, so the sources stay about what a tab holds.
type Fitted struct {
	inner Resolver
	// maxLength bounds the whole: the position is counted against it rather
	// than added on top, so a numbered title is never wider than an unnumbered
	// one.
	maxLength    int
	showPosition bool
}

var _ Resolver = (*Fitted)(nil)

// NewFitted wraps inner, which must leave its names unbounded: the cut here
// is the only one. A width of zero or less takes the default.
func NewFitted(inner Resolver, opts Options) *Fitted {
	if opts.MaxLength <= 0 {
		opts.MaxLength = DefaultMaxLength
	}

	return &Fitted{
		inner:        inner,
		maxLength:    opts.MaxLength,
		showPosition: opts.ShowPosition,
	}
}

// Resolve names the tab, and puts its position in front unless the tab bar is
// too narrow to carry both.
func (f *Fitted) Resolve(tab state.TabState) Decision {
	// The prefix goes in front because the cut takes the tail: a position at
	// the end is the first thing a title too wide for the tab bar would lose.
	prefix := ""
	if f.showPosition {
		prefix = strconv.Itoa(tab.Position) + positionMark
	}

	room := f.maxLength - uniseg.StringWidth(prefix)
	// A title reduced to nothing has lost more than the position is worth.
	if room <= 0 {
		prefix, room = "", f.maxLength
	}

	return f.label(f.inner.Resolve(tab), prefix, room)
}

// ResolvePanes names the panes of a tab and cuts each label to the same width.
// No position leads a pane: the key in front of a title switches to a tab, and
// there is none that switches to a pane.
func (f *Fitted) ResolvePanes(tab state.TabState) []Decision {
	decisions := f.inner.ResolvePanes(tab)
	for i, decision := range decisions {
		decisions[i] = f.label(decision, "", f.maxLength)
	}

	return decisions
}

// label puts prefix in front of what room columns leave of the decision's
// name, and falls back when the cut leaves nothing.
func (f *Fitted) label(decision Decision, prefix string, room int) Decision {
	name := truncate(decision.Name, room)
	if name == "" {
		return genericFallback()
	}

	decision.Name = prefix + name

	return decision
}
