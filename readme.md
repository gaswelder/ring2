# Ring Twice
A home-made mail server

I made it to receive mail notifications from my local Raspberry. For
some reason writing a server seemed simpler (and more fun) than
figuring out how to do what I needed with Exim on Debian.


## Configuration

Example configuration file:

	server {
		hostname pi
		smtp :2525
		pop :11000
		maildir /var/ring2/mail
	}

	lists {
		all
		staff
	}

	users {
		# "123" is the password: gas is cool about local security.
		gas "123" [all]

		# but bob wanted his password replaced with a bcrypt hash
		# (it's "123" too)
		bob $2y$10$SBck8xuRG9QGuANVEnDWXu7s2E6.1hROVch5QS2Ao6yqXLXn1z692 [all, staff]
	}

The `server` section has the following keys:

* `hostname` - host's domain name;
* `smtp` - SMTP listen address;
* `pop` - POP listen address;
* `maildir` - directory where mail will be stored;
* `debug` - if present, server and client commands will be echoed on the standard error output.

The `hostname` should probably be the same as the output of "hostname"
or "uname -n" command. This value affects the addresses of users.

Ring2 is basically two combined servers: POP and SMTP. Both parts may
be enabled, or just one of them. To enable POP, specify the `pop`
parameter, and likewise for SMTP.

Listen addresses have form "[<addr>]:<port>". For example,
"localhost:25" will listen only for local connections on port 25,
whereas ":25" will allows listening for all connections on that port.

The maildir must be writable by the server's process. If it doesn't
exist, the server will try to create it on launch.

The `lists` section defines mailing lists. Users are assigned to mailing
lists in the "users" section.

The `users` section has contains lines describing the users in form:

	name password [lists]

The password part may be a plain password in double quotes or a bcrypt
hash. The lists must be enumerated in square brackets. For example, if
Bob is in lists "all" and "staff", his record might look like:

	bob "bob-rules" [all, staff]
