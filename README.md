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
| `<click:ACTION:VALUE>`                      | `run_command`, `suggest_command`, `open_url`, `open_file`, `suggest_command`, `change_page`, `copy_to_clipboard`, `show_dialog`, `custom` |
| `<hover:show_text:VALUE>`                   | See limitations below                                                                                                                     |
| `<reset>`                                   | Clears all style for the remainder of the current scope; never closes                                                                     |
| `<br>`, `<newline>`                         | Inserts a literal newline; never has children or a close tag                                                                     |

## Testing

```
go test ./...
```
