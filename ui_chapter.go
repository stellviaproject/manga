package main

type UIChapter struct {
	chapter *Chapter
	checked bool
}

func NewUIChapter(chapter *Chapter, checked bool) *UIChapter {
	return &UIChapter{
		chapter: chapter,
		checked: checked,
	}
}

func (ui *UIChapter) ReadChapter() {

}

func (ui *UIChapter) OnChanged(checked bool) {
	ui.checked = checked
}
