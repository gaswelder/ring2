<?php
error_reporting(-1);

$f = stream_socket_client("tcp://localhost:2525");

$c = new tp_client($f);

$c->expect(220);

$c->writeLine("HELO %s", "foo");
$c->expect(250);

$c->writeLine("MAIL FROM:<zilch@foo>");
$c->expect(250);

$c->writeLine("RCPT TO:<joe@yabadoo.com>");
$c->expect(250);

$c->writeLine("DATA");
$c->expect(354);

$c->writeLine("Date: " . date('r'));
$c->writeLine("Subject: test");
$c->writeLine("From: zilch@foo");
$c->writeLine("");
$c->writeLine("Hey you");
$c->writeLine(".");
$c->expect(250);

$c->writeLine("QUIT");
$c->expect(250);


fclose($f);

var_dump($c->err());


/*
 * Transmission protocol reader/writer.
 * Implements the command and response syntax used in *TP protocols
 * such as FTP, POP and SMTP, from the client side.
 */

class tp_client
{
	private $err = null;
	private $conn = null;

	function __construct($conn) {
		$this->conn = $conn;
	}
	
	function err() {
		return $this->err;
	}

	function expect($code) {
		if($this->err) {
			return false;
		}

		$line = fgets($this->conn);
		fwrite(STDERR, "S: $line");

		$n = sscanf($line, "%d %s\r\n", $rcode, $rmsg);
		if($n < 2) {
			$this->err = "Could not read command";
			return false;
		}

		if($rcode != $code) {
			$this->err = "Expected $code, got $rcode ($rmsg)";
			return false;
		}

		return true;
	}

	function writeLine($fmt, $args___ = null)
	{
		if($this->err) {
			return false;
		}

		$args = func_get_args();
		$line = call_user_func_array('sprintf', $args);

		return $this->write($line."\r\n");
	}

	function write($s) {
		if($this->err) {
			return false;
		}
		fwrite(STDERR, "$s");
		return fwrite($this->conn, $s);
	}
}


?>
