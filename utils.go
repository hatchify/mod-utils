package gomu

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	com "github.com/hatchify/mod-common"
)

// ExitWithErrorMessage prints message and exits
func ExitWithErrorMessage(message string) {
	com.Println(message)
	Exit(1)
}

// Exit shows help then exits with status prints message and exits
func Exit(status int) {
	os.Exit(status)
}

// ShowWarningOrQuit will exit if user declines warning
func ShowWarningOrQuit(message string) {
	if !ShowWarning(message) {
		com.Println("Exiting...")
		Exit(0)
	}
}

// ShowWarning prints warning message and waits for user to confirm
func ShowWarning(message string) (ok bool) {
	if com.GetLogLevel() <= com.SILENT {
		// Don't show warnings for silent or name-only
		return true
	}

	var err error
	var text string
	reader := bufio.NewReader(os.Stdin)

	for err == nil {
		if text = strings.TrimSpace(text); len(text) > 0 {
			switch text {
			case "y", "Y", "Yes", "yes", "YES", "ok", "OK", "Ok":
				ok = true
				return
			default:
				com.Println("Nevermind then! :)")
				return
			}
		}

		// No newline. name-only already exited above
		fmt.Print(message + " [y|yes|ok]: ")
		text, err = reader.ReadString('\n')
	}

	com.Println("Oops... Something went wrong.")
	return
}
