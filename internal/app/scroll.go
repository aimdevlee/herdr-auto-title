package app

import (
	"maps"
	"sync"

	"github.com/kryptamine/herdr-auto-title/internal/resolver"
)

// scroller remembers, per tab and per pane, how far the name it was last given
// has slid. The tick moves once per poll, so the poll interval is the sliding
// speed.
type scroller struct {
	mu sync.Mutex
	// step is how many columns a name moves per poll.
	step   int
	slides map[string]sliding
}

type sliding struct {
	name string
	tick int
	// moved says the last fit showed this name further along than its head.
	moved bool
}

func newScroller(step int) *scroller {
	return &scroller{step: step, slides: make(map[string]sliding)}
}

// Fit slides a name too wide for width, starting over whenever the name
// itself changes.
func (s *scroller) Fit(id, name string, width int) string {
	s.mu.Lock()
	defer s.mu.Unlock()

	current := s.slides[id]
	if current.name == name {
		current.tick++
	} else {
		current = sliding{name: name}
	}

	shown := resolver.Slide(name, width, current.tick, s.step)
	current.moved = current.tick > 0 && shown != name
	s.slides[id] = current

	return shown
}

// sliding reports whether the label is the one it was last given, moved along,
// rather than a new one.
func (s *scroller) sliding(id string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	return s.slides[id].moved
}

// Retain forgets the tabs and panes the session no longer holds.
func (s *scroller) Retain(tabs, panes map[string]string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	maps.DeleteFunc(s.slides, func(id string, _ sliding) bool {
		_, tab := tabs[id]
		_, pane := panes[id]

		return !tab && !pane
	})
}
