package tui

// ScrollState manages cursor position and scroll offset for a scrollable list.
type ScrollState struct {
	Cursor      int
	Offset      int
	VisibleRows int
}

func (s *ScrollState) Up() bool {
	if s.Cursor <= 0 {
		return false
	}
	s.Cursor--
	if s.Cursor < s.Offset {
		s.Offset = s.Cursor
	}
	return true
}

func (s *ScrollState) Down(itemCount int) bool {
	if s.Cursor >= itemCount-1 {
		return false
	}
	s.Cursor++
	if s.VisibleRows > 0 && s.Cursor >= s.Offset+s.VisibleRows {
		s.Offset = s.Cursor - s.VisibleRows + 1
	}
	return true
}

func (s *ScrollState) First() {
	s.Cursor = 0
	s.Offset = 0
}

func (s *ScrollState) Last(itemCount int) {
	if itemCount <= 0 {
		return
	}
	s.Cursor = itemCount - 1
	if s.VisibleRows > 0 && s.Cursor >= s.VisibleRows {
		s.Offset = s.Cursor - s.VisibleRows + 1
	} else {
		s.Offset = 0
	}
}

func (s *ScrollState) VisibleRange(itemCount int) (start, end int) {
	start = s.Offset
	end = s.Offset + s.VisibleRows
	if end > itemCount {
		end = itemCount
	}
	if start > end {
		start = end
	}
	return start, end
}

func (s *ScrollState) ClampToCount(itemCount int) {
	if s.Cursor >= itemCount {
		s.Cursor = itemCount - 1
	}
	if s.Cursor < 0 {
		s.Cursor = 0
	}
	if s.Offset > s.Cursor {
		s.Offset = s.Cursor
	}
	if s.Offset < 0 {
		s.Offset = 0
	}
}

func (s *ScrollState) Reset() {
	s.Cursor = 0
	s.Offset = 0
}
