package micromessage

// Converts a minimessage string to a text component string
func Parse(str string) string {
	c, err := Deserialize(str)
	if err != nil {
		return str
	}
	return PlainText(c)
}
