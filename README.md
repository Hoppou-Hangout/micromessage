# micromessage

A MiniMessage parser and renderer for Go, targeting Minekube's component
model (`go.minekube.com/common/minecraft/component`).

```go
msg, err := micromessage.Deserialize("<gradient:aqua:blue>Welcome</gradient> <red><bold>friend</bold></red>!")
if err != nil {
    return err
}
return player.SendMessage(msg)
```

## Installation

```
go get github.com/Hoppou-Hangout/micromessage
```

## Usage

```go
package main

import (
    "github.com/Hoppou-Hangout/micromessage"
)

func main() {
    msg, err := micromessage.Deserialize("<red>Hello <bold>world</bold></red>!")
    if err != nil {
        panic(err)
    }
    // msg is a *component.Text from go.minekube.com/common/minecraft/component.
    // Pass it directly to Gate:
    //   player.SendMessage(msg)
}
```

`MustDeserialize` is also available for static messages defined at init
time, where a parse error should just panic instead of being handled:

```go
var welcomeMsg = micromessage.MustDeserialize("<gradient:gold:yellow>Welcome!</gradient>")
```

## Supported tags

| Tag                                         | Notes                                                                                                                                     |
| ------------------------------------------- | ----------------------------------------------------------------------------------------------------------------------------------------- |
| `<red>`, `<blue>`, etc.                     | All 16 legacy named colors, plus `<grey>`/`<dark_grey>` British aliases                                                                   |
| `<#rrggbb>`                                 | Bare hex color                                                                                                                            |
| `<color:NAME>`, `<colour:NAME>`, `<c:NAME>` | Explicit color tag, name or hex                                                                                                           |
| `<bold>`, `<b>`                             | and `<italic>`/`<i>`/`<em>`, `<underlined>`/`<u>`, `<strikethrough>`/`<st>`, `<obfuscated>`/`<obf>`                                       |
| `<!bold>`                                   | Negation shorthand, turns a decoration off                                                                                                |
| `<bold:false>`                              | Explicit boolean form, same effect as `<!bold>`                                                                                           |
| `<gradient>`                                | Defaults to white to black with no arguments                                                                                              |
| `<gradient:c1:c2:...:cN>`                   | Any number of color stops                                                                                                                 |
| `<gradient:c1:c2:phase>`                    | Trailing numeric argument shifts the gradient's starting point                                                                            |
| `<rainbow>`, `<rainbow:phase>`, `<rainbow:!>`| Hue cycles across the wrapped text; `!` reverses direction, phase is in tenths                                                            |
| `<transition:c1:c2:...:cN[:phase]>`         | Same args as `<gradient>`, but (matching real MiniMessage) resolves to a single static color, not a per-character blend                   |
| `<shadow:NAME_OR_HEX:[alpha]>`, `<shadow:#RRGGBBAA>`, `<!shadow>` | Text shadow color; alpha (0-1) defaults to 0.25                                                                     |
| `<insert:VALUE>`                            | Shift-click insertion text                                                                                                                |
| `<font:KEY>`, `<font:NAMESPACE:KEY>`        | Sets the font resource key (default namespace `minecraft`)                                                                                |
| `<click:ACTION:VALUE>`                      | `run_command`, `suggest_command`, `open_url`, `open_file`, `suggest_command`, `change_page`, `copy_to_clipboard`, `show_dialog`, `custom` |
| `<hover:show_text:VALUE>`                   | `VALUE` is itself parsed as MiniMessage, so it can carry its own colors/formatting                                                        |
| `<hover:show_item:ID[:COUNT[:NBT]]>`        | `ID` needs quoting if namespaced, e.g. `"minecraft:diamond"`; bare `ID` defaults to the `minecraft` namespace                              |
| `<hover:show_entity:TYPE:UUID[:NAME]>`      | Same `TYPE` quoting rule as `show_item`; `NAME` is parsed as MiniMessage                                                                   |
| `<lang:KEY[:with...]>` (`tr`, `translate`)  | Translatable component; each `with` argument is itself parsed as MiniMessage                                                             |
| `<lang_or:KEY:FALLBACK[:with...]>` (`tr_or`, `translate_or`) | Same as `<lang>`, with a client-side fallback string                                                                     |
| `<reset>`                                   | Clears all style for the remainder of the current scope; never closes                                                                     |
| `<br>`, `<newline>`                         | Inserts a literal newline; never has children or a close tag                                                                              |

Not implemented: `<key>` (keybind), `<selector>`, `<score>`, `<nbt>`/`<data>` — the underlying
`go.minekube.com/common` component model has no component types for these. `<pride>`, `<sprite>`,
`<head>` are also unimplemented (client-rendered visuals, not representable as plain text/color).

## API

`Deserialize`/`MustDeserialize` take optional `Option`s, mirroring Adventure's
`MiniMessage.builder()` API:

```go
msg, err := micromessage.Deserialize(
    "Hello <name/>, you have <score:42/> points! <red>bold isn't here</red>",
    micromessage.WithTagResolver(micromessage.Placeholder("name", micromessage.Parsed("<gold>Tom</gold>"))),
    micromessage.WithTagResolver(scoreResolver),
)
```

By default every standard tag is recognized (`StandardTags.All()`, equivalent to Adventure's
`MiniMessage.miniMessage()`) and anything that doesn't resolve — an unknown tag name, or a
built-in tag you've excluded — is rendered as literal text instead of erroring, matching
Adventure's default (non-strict) error handling.

### Tag resolvers

All tag resolution goes through a `TagResolver`:

```go
type TagResolver interface {
    ResolveTag(name string, args *ArgumentQueue) (Tag, bool)
}
```

`name` is already lower-cased. `ArgumentQueue` gives access to the tag's raw arguments —
`Pop()`/`PopOr(msg)`/`PeekOr(msg)`/`Rest()`/`HasNext()`; `PopOr`/`PeekOr` abort resolution with
an error (surfaced from `Deserialize`) if a required argument is missing, matching Adventure's
`Tag.Argument#popOr`.

- `Placeholder(name, tag)` matches one tag name case-insensitively regardless of arguments —
  the common case for a single named value.
- `TagResolverFunc` wraps a plain function for tags that need their arguments, e.g.
  `<score:42>` or `<selector:@a>`.
- `NewTagResolverBuilder().Resolver(a).Resolver(b).Build()` composes several resolvers into
  one, tried in order — mirrors `TagResolver.builder()`.
- `WithTagResolver` adds a resolver on top of the active tag set (the common case: add a
  placeholder without touching any built-in tag). `WithTags` *replaces* the active set —
  built-in tags not included (directly, or via `StandardTags.All()`) stop being recognized:

  ```go
  // Only "<red>"/"<color:...>" are recognized; "<bold>" renders as literal text.
  msg, err := micromessage.Deserialize("<green><bold>Hai",
      micromessage.WithTags(micromessage.StandardTags.Color()))
  ```

  `StandardTags` exposes each built-in category individually — `Color`, `Decorations`,
  `Click`, `HoverEvent`, `Insertion`, `Font`, `Shadow`, `Gradient`, `Rainbow`, `Transition`,
  `Translatable`, `Newline`, `Reset`, and `All` — matching Adventure's `StandardTags` class.

### Tags

A `TagResolver` returns a `Tag`, built with exactly one of:

- `Text(value)` — literal text, taking the ambient style but never re-parsed. Self-closing by
  default. Matches `Placeholder.unparsed`.
- `Parsed(value)` — parsed as its own MiniMessage snippet and spliced in, inheriting the
  ambient style; the same resolvers apply recursively inside it (capped at 64 levels deep, to
  catch self-referential placeholders). A pragmatic stand-in for Adventure's `PreProcess` tags
  — see the doc comment on `Parsed` for how it differs.
- `ComponentTag(comp)` — a pre-built `component.Component` inserted verbatim. Self-closing by
  default. Matches `Placeholder.component`.
- `StylingTag(styles...)` — wraps its content, applying style changes built from `ColorStyle`,
  `DecorationStyle`, `ClickStyle`, `HoverStyle`, `FontStyle`, `InsertionStyle`, `ShadowStyle`.
  Matches `Tag.styling(...)`; this is how every built-in styling tag (`<color>`, `<bold>`,
  `<click>`, ...) is itself implemented.
- `ModifyingTagValue(m)` — a custom `ModifyingTag` (`Visit`/`PostVisit`/`Apply`), for tags that
  need to see and transform their whole rendered subtree, the way `<gradient>`/`<rainbow>` do.
- `DirectiveTag(Reset)` — a `ParserDirective`: the tag behaves exactly like `<reset>`, closing
  every currently open tag. Register it under another name with `Resolver(name, Reset)`:

  ```go
  clearTag := micromessage.Resolver("clear", micromessage.Reset)
  msg, err := micromessage.Deserialize("<red>hello <bold>world<clear>, how are you?",
      micromessage.WithTagResolver(clearTag))
  ```

Call `.SelfClosing()` on any `Tag` to mark it as never taking a close tag or wrapped content
(only `Text`/`ComponentTag` default to this).

### Presets

`DefaultPreset`, `NonInteractablePreset`, and `FormattedTextPreset` mirror Adventure's
`MiniMessage.Preset`s — pass `SomePreset.Apply()` as an `Option`:

```go
msg, err := micromessage.Deserialize(input, micromessage.NonInteractablePreset.Apply())
```

`NonInteractablePreset` drops `<click>`/`<hover>`/`<insert>` from the tag set and strips any
interactable style a custom resolver still introduces. `FormattedTextPreset` additionally drops
any non-text component (e.g. a `<lang>` translation) from the result.

### Strict mode and debugging

`WithStrict(true)` makes an unclosed tag a parse error instead of auto-closing at EOF, matching
`Builder#strict(true)` — tags that simply don't resolve to anything are still rendered as
literal text either way. `WithDebug(fn)` registers a callback for diagnostic messages.

### Preprocessors

`WithPreprocessor` registers a function that rewrites the raw input string before it's lexed
and parsed, e.g. to translate legacy `&`-formatted color codes into MiniMessage tags:

```go
msg, err := micromessage.Deserialize(
    "&cHello &lworld",
    micromessage.WithPreprocessor(func(s string) string {
        return strings.NewReplacer("&c", "<red>", "&l", "<bold>").Replace(s)
    }),
)
```

Multiple preprocessors run in the order added, each seeing the previous one's output.

## Testing

```
go test ./...
```
