package pop

import "testing"

func TestParsing(t *testing.T) {
	t.Run("valid command", func(t *testing.T) {
		cmd, err := parseCommand("foo bar bar bar\r\n")
		if err != nil {
			t.Fatal(err)
		}
		if cmd.name != "foo" {
			t.Fatal("foo expected as the command")
		}
		if cmd.arg != "bar bar bar" {
			t.Fatal("bar bar bar expected as the argument")
		}
	})

	t.Run("invalid command", func(t *testing.T) {
		cmd, err := parseCommand("300-hey-there dude")
		if err == nil {
			t.Fatalf("expected to fail, got %v", cmd)
		}
	})
}
