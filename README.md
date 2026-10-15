# vex

A small terminal editor for XML files, built with
[Bubble Tea](https://github.com/charmbracelet/bubbletea) and
[Lip Gloss](https://github.com/charmbracelet/lipgloss).

## Install

```sh
go install github.com/ixexel661/vex@latest
```

From a checkout, `go install .` does the same. The binary lands in `$GOBIN`
(default `~/go/bin`), which has to be in your `PATH`.

## Usage

```sh
vex file.xml
```

vex opens the file in normal mode with syntax highlighting and line numbers.
Press `i` to edit and `esc` to stop editing. A file that does not exist yet is
created on the first save. The status bar shows whether the document is
well-formed XML and, if not, the line of the first error.

| Mode   | Keys                                  | Action                                    |
|--------|---------------------------------------|-------------------------------------------|
| normal | `i`                                   | start editing                             |
| normal | `f`                                   | format (re-indent) the document           |
| normal | `q`, `ctrl+c`                         | quit, asks first if there are changes     |
| both   | `ctrl+s`                              | save                                      |
| both   | `ctrl+f`                              | search                                    |
| both   | arrows, `home`/`end`, `pgup`/`pgdown` | move                                      |
| both   | `ctrl+home`/`ctrl+end`                | go to the start/end of the document       |
| insert | `esc`, `ctrl+c`                       | stop editing                              |
| insert | `enter`, `backspace`, `delete`, `tab` | new line (keeps indentation), delete, indent |

Formatting only changes whitespace between tags. Elements that contain text
are kept exactly as they are.

### Search

`ctrl+f` opens the search bar in place of the message line. vex searches while
you type, in one of two modes:

- **Text** finds the exact text. The cursor jumps to the first match from
  where it was, and all matches are marked in the document.
- **Fuzzy** finds lines that contain the typed characters in order, e.g. `ttl`
  finds `<title>`. The best matching lines are listed above the status bar.

Case only matters if the query contains an upper case letter. The next search
starts with the last query and mode: typing replaces the query, `up`/`down`
and `enter` use it.

| Keys            | Action                                    |
|-----------------|-------------------------------------------|
| `up`/`down`     | previous/next match                       |
| `enter`         | close the search, the cursor stays        |
| `esc`, `ctrl+c` | cancel, the cursor goes back              |
| `tab`           | switch between text and fuzzy search      |

## Project layout

```
main.go                 wires everything together
internal/
  app/                  Bubble Tea model: keys -> commands, layout, dialogs
  command/              named commands and the Context they run against
  keymap/               key bindings per mode
  mode/                 normal and insert mode
  editor/               cursor, scrolling, editing and rendering
  document/             reading and writing files
  buffer/               the text itself
  highlight/            highlighting primitives and a per-line cache
  search/               text and fuzzy search
  filetype/             file type registry
    xml/                XML highlighter, formatter and validator
  config/, theme/       settings and colors
```

## Extending vex

**A new command:** write a `func(ctx command.Context) error` in
`internal/command`, add it to the list of its topic (e.g.
`documentCommands()`), and bind a key in `keymap.Default()`. Give it a
`Title` to show it in the key hints.

**A new file format:** create `internal/filetype/<format>` with a function
returning a `filetype.Type`. Only fill in what the format supports
(`Highlighter`, `Formatter`, `Validator`), then register it in `main.go`.

**A new search mode:** implement `search.Matcher` in `internal/search`, e.g.
for regular expressions or XPath, and add it to the matchers in `main.go`.
`tab` switches through them in that order. The matches of a matcher that
ranks them are listed, the others are marked in the text.

**Plugins (planned):** commands, key bindings, file types and search modes are
all set up in `main.go`. A plugin loader can extend them there and use
`command.Context` as its API.
