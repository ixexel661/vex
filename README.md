# vex

A small terminal editor for XML files, built with
[Bubble Tea](https://github.com/charmbracelet/bubbletea) and
[Lip Gloss](https://github.com/charmbracelet/lipgloss).

## Install

Every [release](https://github.com/ixexel661/vex/releases) has a binary for
Linux, macOS and Windows on x64 and ARM64: `vex-linux-amd64`,
`vex-linux-arm64`, `vex-darwin-amd64`, `vex-darwin-arm64`,
`vex-windows-amd64.exe` and `vex-windows-arm64.exe`. `SHA256SUMS` lists their
checksums. Rename the binary to `vex` (`vex.exe` on Windows), put it in a
directory in your `PATH` and, on Linux and macOS, make it executable with
`chmod +x vex`. The binaries are not signed, so after a download in the
browser, macOS needs `xattr -d com.apple.quarantine vex` before it runs it.

With Go installed:

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

```sh
vex new file.xml
vex new
```

`vex new file.xml` creates the file with an XML declaration and opens it in
insert mode below it. A name without an extension gets `.xml`. If the file
exists already, vex asks before it overwrites it. `vex new` without a name
starts an `untitled` document and asks for the name when you save it. To
open a file called `new`, use `vex ./new`.

| Mode   | Keys                                  | Action                                    |
|--------|---------------------------------------|-------------------------------------------|
| normal | `i`                                   | start editing                             |
| normal | `f`                                   | format (re-indent) the document           |
| normal | `s`                                   | settings                                  |
| normal | `q`, `ctrl+c`                         | quit, asks first if there are changes     |
| both   | `ctrl+s`                              | save                                      |
| both   | `ctrl+f`                              | search                                    |
| both   | arrows, `home`/`end`, `pgup`/`pgdown` | move                                      |
| both   | `ctrl+home`/`ctrl+end`                | go to the start/end of the document       |
| insert | `esc`, `ctrl+c`                       | stop editing                              |
| insert | `enter`, `backspace`, `delete`        | new line (keeps indentation), delete      |
| insert | `tab`, `shift+tab`                    | indent to the next stop, outdent the line |
| insert | `>`                                   | end a start tag and add its end tag       |

Formatting changes the whitespace between tags and, if the settings ask for
it, wraps long tags and shortens empty elements. Text is kept exactly as it
is.

Indenting while typing follows the `indent` setting: `tab` fills up to the
next indent stop, `shift+tab` takes one level off the start of the line, and
`backspace` in the indentation deletes back to the previous stop.

In insert mode, the start and end tag of an element stay in step. With the
cursor on a tag name, both names are marked, and typing or deleting in one
of them changes the other one too. Characters that cannot be part of a name,
like the space before an attribute, only go where you type them. Typing `>`
at the end of a start tag adds its end tag right after it, with the cursor
in between, unless the element has an end tag already.

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

### Settings

`s` opens the settings over the text. Every change shows at once: the
format settings format the document right away, always the text from when
the settings opened. `esc` brings that text back, `enter` keeps the
formatted one, which is then a change to save.

| Keys              | Action                                  |
|-------------------|-----------------------------------------|
| `up`/`down`       | select a setting                        |
| `left`/`right`    | change it                               |
| `enter`           | save the settings to the config file    |
| `esc`, `ctrl+c`   | cancel, the settings go back            |

The bottom line of the settings shows the version of vex and where the
config file is: `~/.config/vex/config.toml` (on macOS
`~/Library/Application Support/vex`, on Windows `%AppData%\vex`). vex
creates it on the first start, and you can also change it by hand:

```toml
theme = "basic"
tab-width = 4                # columns of a tab character

[format]
indent = "2 spaces"          # for formatting and the tab key, or "tabs"
self-close-empty = false     # <a></a> becomes <a/>
keep-blank-lines = false     # single blank lines between elements stay
wrap-tags-longer-than = 0    # each attribute on its own line, 0 = never
```

Saving from the settings rewrites the file, so comments in it get lost. If
the file has a problem, vex says so when it starts and uses the defaults
for what it cannot read. The font is the one of your terminal.

### Themes

The colors come from the themes in `~/.config/vex/themes`. vex puts
`basic.toml` there, which uses the 16 colors of the terminal, and the
settings offer every `.toml` file in that directory. A theme has a
`[syntax]` table for highlighted text and a `[ui]` table for everything
around it:

```toml
[syntax]
tag = { fg = "blue" }
comment = { fg = "bright-black", italic = true }

[ui]
status-bar = { fg = "#ffffff", bg = "238" }
```

A style can have a foreground (`fg`) and a background (`bg`) color and be
`bold`, `italic` or `underline`. A color is a name (`black`, `red`,
`green`, `yellow`, `blue`, `magenta`, `cyan`, `white`, each also with
`bright-` in front), a number from 0 to 255 or a hex color. For a theme of
your own, copy `basic.toml` and change it. It only needs the styles it
changes, the others come from the basic theme built into vex.

## Releases

Pushing a tag that starts with `v` builds a release:

```sh
git tag v1.0.0
git push origin v1.0.0
```

The workflow in `.github/workflows/release.yml` runs the tests, builds the
six binaries and publishes them with their checksums as the release of the
tag. Tags with a `-`, like `v1.1.0-beta`, become pre-releases. The settings
show the version of the tag, builds from a checkout show `dev`.

## Project layout

```
main.go                 wires everything together
new.go                  the arguments, and the files vex new creates
version.go              the version the settings show
.github/workflows/      the release build
themes/                 themes built into vex
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
    xml/                XML highlighter, formatter, validator and tag pairs
  config/               settings and the config file
  theme/                styles and theme files
```

## Extending vex

**A new command:** write a `func(ctx command.Context) error` in
`internal/command`, add it to the list of its topic (e.g.
`documentCommands()`), and bind a key in `keymap.Default()`. Give it a
`Title` to show it in the key hints.

**A new file format:** create `internal/filetype/<format>` with a function
returning a `filetype.Type`. Only fill in what the format supports
(`Template`, `Highlighter`, `Formatter`, `Validator`, `Tags`), then register
it in `main.go`. The first registered type is the one of `vex new`.

**A new search mode:** implement `search.Matcher` in `internal/search`, e.g.
for regular expressions or XPath, and add it to the matchers in `main.go`.
`tab` switches through them in that order. The matches of a matcher that
ranks them are listed, the others are marked in the text.

**A new setting:** add a field with a `toml` tag to `config.Config` and its
default to `config.Default()`. To change it in the settings, add a row to
`settingRows()` in `internal/app/settings.go`.

**A new style:** add a field to `theme.Theme` (or a kind to `highlight`),
its key to `internal/theme/keys.go` and its colors to `themes/basic.toml`.
A test fails as long as `basic.toml` misses a style.

**Plugins (planned):** commands, key bindings, file types and search modes are
all set up in `main.go`. A plugin loader can extend them there and use
`command.Context` as its API.
