package command

import "github.com/ixexel661/vex/internal/editor"

func cursorCommands() []Command {
	return []Command{
		{Name: "cursor.up", Run: onEditor((*editor.Editor).MoveUp)},
		{Name: "cursor.down", Run: onEditor((*editor.Editor).MoveDown)},
		{Name: "cursor.left", Run: onEditor((*editor.Editor).MoveLeft)},
		{Name: "cursor.right", Run: onEditor((*editor.Editor).MoveRight)},
		{Name: "cursor.line-start", Run: onEditor((*editor.Editor).MoveLineStart)},
		{Name: "cursor.line-end", Run: onEditor((*editor.Editor).MoveLineEnd)},
		{Name: "cursor.page-up", Run: onEditor((*editor.Editor).PageUp)},
		{Name: "cursor.page-down", Run: onEditor((*editor.Editor).PageDown)},
		{Name: "cursor.top", Run: onEditor((*editor.Editor).MoveTop)},
		{Name: "cursor.bottom", Run: onEditor((*editor.Editor).MoveBottom)},
		{Name: "cursor.next-tag", Run: onEditor((*editor.Editor).MoveToNextTag)},
		{Name: "cursor.prev-tag", Run: onEditor((*editor.Editor).MoveToPrevTag)},
		{Name: "cursor.parent", Run: onEditor((*editor.Editor).MoveToParent)},
		{Name: "cursor.first-child", Run: onEditor((*editor.Editor).MoveToFirstChild)},
	}
}
