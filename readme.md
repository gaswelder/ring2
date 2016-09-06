# Ring Twice
A home-made mail server


## Why?

To receive mail notifications from my Raspberry. For some reason
writing a server was more fun than figuring out how to do that with
Exim on Debian.


## Configuration

The `server` section has the following keys:

* `hostname` - host's domain name;
* `smtp` - SMTP listen address;
* `pop` - POP listen address;
* `maildir` - directory where mail will be stored.

The `hostname` should probably be the same as the output of "hostname"
or "uname -n" commands. This value affects the addresses of users.

Ring2 is basically two combined servers: POP and SMTP. Both parts may
be enabled, or just one of them. To enable POP, specify the `pop`
parameter, and likewise for SMTP.

Listen addresses have form "[<addr>]:<port>". For example,
"localhost:25" will listen only for local connections on port 25,
whereas ":25" will allows listening for all connections on that port.

The maildir must be writable by the server's process. If it doesn't
exist, the server will try to create it on launch.
