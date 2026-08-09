package micromessage_test

import (
	"fmt"

	"github.com/Hoppou-Hangout/micromessage"
)

// This mirrors how you'd use the library inside a Gate plugin:
//
//	msg, err := micromessage.Deserialize("<gradient:aqua:blue>Welcome</gradient> <red>%s</red>!")
//	if err != nil {
//	    return err
//	}
//	return player.SendMessage(msg) // player is a go.minekube.com/gate proxy.Player
//
// player.SendMessage takes a component.Component (msg component.Component, opts
// ...command.MessageOption), and *component.Text (what Deserialize returns)
// implements that interface directly, so no adapting is needed.
func Example() {
	msg, err := micromessage.Deserialize(`<gradient:aqua:blue>Welcome</gradient> <red><bold>friend</bold></red>!`)
	if err != nil {
		panic(err)
	}
	fmt.Println(len(msg.Children()) > 0)
	// Output: true
}
