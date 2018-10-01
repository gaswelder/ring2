package smtp

// func sendBounce(fpath, rpath *Path, config *Config) error {
// 	var b bytes.Buffer
// 	fmt.Fprintf(&b, "Date: %s\r\n", formatDate())
// 	fmt.Fprintf(&b, "From: ring2@%s\r\n", config.Hostname)
// 	fmt.Fprintf(&b, "To: %s\r\n", rpath.Addr.Format())
// 	fmt.Fprintf(&b, "Subject: mail delivery failure\r\n")
// 	fmt.Fprintf(&b, "\r\n")
// 	fmt.Fprintf(&b, "Sorry, your mail could not be delivered to %s", fpath.Addr.Format())
// 	/*
// 	 * Specify null as reverse-path to prevent loops
// 	 */
// 	return sendMail(b.String(), rpath, nil, config)
// }

// /*
//  * Send a message using SMTP protocol
//  */
// func sendMail(text string, fpath, rpath *Path, config *Config) error {

// 	if len(fpath.Hosts) == 0 {
// 		return errors.New("Empty forward-path")
// 	}

// 	conn, err := net.Dial("tcp", fpath.Hosts[0])
// 	if err != nil {
// 		return err
// 	}

// 	w := newTpClient(conn)
// 	w.Expect(250)

// 	w.WriteLine("HELO %s", config.Hostname)
// 	w.Expect(250)

// 	if rpath != nil {
// 		w.WriteLine("MAIL FROM:" + rpath.Format())
// 	} else {
// 		w.WriteLine("MAIL FROM:<>")
// 	}
// 	w.Expect(250)

// 	w.WriteLine("RCPT TO:" + fpath.Format())
// 	w.Expect(250)

// 	w.WriteLine("DATA")
// 	w.Expect(354)

// 	lines := strings.Split(text, "\r\n")
// 	for _, line := range lines {
// 		/*
// 		 * If the first character of a line is a dot,
// 		 * insert one additional dot.
// 		 */
// 		if len(line) > 0 && line[0] == '.' {
// 			w.Write(".")
// 		}
// 		w.WriteLine(line)
// 	}

// 	w.WriteLine(".")
// 	w.Expect(250)

// 	w.WriteLine("QUIT")
// 	w.Expect(221)

// 	conn.Close()

// 	return w.Err()
// }
