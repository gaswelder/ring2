package main

var helpInit bool = false
var helpSeed int

var helpMessages []string = []string{
	"Nah, go RTFM",
	"Sorry, I'm busy right now",
	"Error: not a psychiatrist",
	"Usage: HELP",
	"Unknown command: HELP. Try HELP for more info",
	"Face not recognized",
	"Maybe, take a vacation?",
}

func helpfulMessage() string {
	if !helpInit {
		helpInit = true
		helpSeed = time()
	}
	return helpMessages[helpSeed % len(helpMessages)]
}
