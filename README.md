# micromessage

A Go implementation of Adventure's [MiniMessage](https://docs.papermc.io/adventure/minimessage/format/)
text format.

## Install

```sh
go get github.com/Hoppou-Hangout/micromessage
```

## Usage

```go
package main

import (
	"fmt"
	"micromessage"
)

func main() {
	comp, err := micromessage.Deserialize("<red>Hello, <b>World</b></red>")
	if err != nil {
		panic(err)
	}

	fmt.Println(micromessage.PlainText(comp)) // "Hello, World"
	fmt.Println(micromessage.Serialize(comp)) // "<red>Hello, <bold>World"
}
```

### Placeholders

```go
comp, _ := micromessage.Deserialize(
	"Hello, <name>!",
	micromessage.StringPlaceholder("name", "World"),
)
```

### Strip / escape tags

```go
micromessage.StripTags("<red>Hello!</red>")  // "Hello!"
micromessage.EscapeTags("<red>Hello!</red>") // "\\<red>Hello!\\</red>"
```

### Strict mode

`Deserialize` degrades gracefully on malformed/unknown tags (rendering them as
literal text). Use `DeserializeStrict` to get a `*ParseError` instead.

## Supported tags

- Colors: named colors (`<red>`, `<gray>`, aliases like `<grey>`), hex
  (`<#ff00ff>`), and `<color:...>` / `<colour:...>` / `<c:...>`
- Decorations: `<bold>`, `<italic>`, `<underlined>`, `<strikethrough>`,
  `<obfuscated>` (and shorthands `<b>`, `<i>`/`<em>`, `<u>`, `<st>`, `<obf>`),
  plus negation (`<!bold>`) and explicit state (`<bold:false>`)
- `<gradient>` and `<rainbow>` (with phase/reverse arguments)
- `<click:...>`, `<hover:show_text:...>`, `<insert:...>`, `<font:...>`
- `<reset>`, `<newline>` / `<br>`

## Tests

```sh
go test ./...
```
