package cfg

import (
	"fmt"
	"io/ioutil"
	"strings"
)

type section map[string]string
type config map[string]section

// ParseFile parses the config file at the given path.
func ParseFile(path string) (cfg config, err error) {
	src, err := ioutil.ReadFile(path)
	if err != nil {
		return
	}
	return parseString(string(src))
}

// parseString parses a config file given as a string.
func parseString(s string) (config, error) {
	text := stripComments(s)

	cfg := make(config)

	b := newScanner(text)
	for {
		b.skipSpaces()
		if !b.more() {
			break
		}

		var name string
		var sec section
		name, sec, err := readSection(b)
		if err != nil {
			return nil, err
		}
		cfg[name] = sec
	}
	return cfg, nil
}

// Removes comments from the text
func stripComments(s string) string {
	lines := strings.Split(s, "\n")
	for k, line := range lines {
		pos := strings.Index(line, "#")
		if pos >= 0 {
			lines[k] = lines[k][:pos]
		}
	}
	return strings.Join(lines, "\n")
}

func readSection(b *scanner) (name string, sec section, err error) {

	name = readName(b)
	if name == "" {
		err = fmt.Errorf("Identifier expected")
		return
	}

	b.skipSpaces()
	if !b.more() {
		return
	}
	b.expect('{')
	b.skipSpaces()

	sec = make(section)

	for b.next() != '}' {
		key := readName(b)
		if key == "" {
			err = fmt.Errorf("property expected here: [...]%s[...]", b.restSnippet())
			return
		}

		for b.next() == ' ' || b.next() == '\t' {
			b.get()
		}

		val := ""
		for b.next() != '\n' && b.next() != '\r' {
			val += string(b.get())
		}

		// If this is a key without a value, put something non-empty there
		// so that its presence can be checked as cfg[section][key] != "".
		if val == "" {
			val = "true"
		}

		sec[key] = val

		b.skipSpaces()
	}
	b.expect('}')
	err = b.err
	return
}

func isAlpha(ch byte) bool {
	return (ch >= 'a' && ch <= 'z') || (ch >= 'A' && ch <= 'Z')
}

func isDigit(ch byte) bool {
	return ch >= '0' && ch <= '9'
}

func readName(b *scanner) string {
	name := ""
	if !isAlpha(b.next()) {
		return name
	}

	for isAlpha(b.next()) || isDigit(b.next()) || b.next() == '-' {
		name += string(b.get())
	}
	return name
}
