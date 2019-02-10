# The conf package

This package parses configuration files formatted like this:

	# comment
	server {
		listen :2525
		relay
		hostname home.loc # comment
	}

	users {
		bob
		alice
		jack
	}

The result of parsing is a two-level map. For the example above, if `m`
is the resulting map, `m["server"]["listen"]` would be equal to
`":2525"`, and `m["users"]["bob"]` would be equal to `""`.

This format is practically the same as "ini", just with a different
syntax. I just like curly braces more...
