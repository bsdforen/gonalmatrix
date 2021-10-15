package main

import (
	"fmt"
	"io"
	"log"
	"strings"
	"time"

	"maunium.net/go/mautrix"
	"maunium.net/go/mautrix/event"
)

// ----

// Client instance, representing the server connection.
var matrixClient *mautrix.Client

// Logger instance, logs events into a file.
var matrixLogger *log.Logger

// ----

// Wrapper to log message events to a file.
func matrixLogMessageEvent(evt *event.Event, message string) {
	timestr := time.Now().Format("2006/01/02 03:04:05")
	matrixLogger.Printf("[%v] %v -> %v: %v\n", timestr, evt.RoomID, evt.Sender, message)
}

// Wrapper to print nicely formated actions to stdout.
func matrixPrintAction(evt *event.Event, action string) {
	timestr := time.Now().Format("2006/01/02 03:04:05")
	fmt.Printf(" * [%v] %v -> %v: %v\n", timestr, evt.RoomID, evt.Sender, action)
}

// Wrapper to print an error to stdout.
func matrixPrintError(evt *event.Event, err error) {
	timestr := time.Now().Format("2006/01/02 03:04:05")
	fmt.Printf(" * [%v] %v -> %v: %v\n", timestr, evt.RoomID, evt.Sender, err)
}

// Handles message events.
func matrixHandleMessageEvent(source mautrix.EventSource, evt *event.Event) {
	message := evt.Content.AsMessage().Body

	// Log everything that has a body.
	if len(message) != 0 {
		matrixLogMessageEvent(evt, message)
	}

	// !info -> Answer with factoid.
	if strings.HasPrefix(message, "!info") {
		split := strings.SplitN(message, " ", 2)

		if len(split) == 1 {
			// No argument -> random factoid.
			matrixPrintAction(evt, "!info")

			fact, err := sqliteFactoidGetRandom()
			if err != nil {
				matrixPrintError(evt, err)
			}
			matrixClient.SendText(evt.RoomID, fact)
		} else if len(split) == 2 {
			// Argument -> all factoids with that key.
			key := strings.Trim(split[1], " ")
			matrixPrintAction(evt, fmt.Sprintf("!info %v", key))

			fact, err := sqliteFactoidGetForKey(key)
			if err != nil {
				matrixPrintError(evt, err)
			}
			matrixClient.SendText(evt.RoomID, fact)
		}
	}

	// !ping -> Anwer with 'pong!'.
	if strings.HasPrefix(message, "!ping") {
		matrixPrintAction(evt, "!ping")
		matrixClient.SendText(evt.RoomID, "pong!")
	}

	// !version -> Answer with the version numer.
	if strings.HasPrefix(message, "!version") {
		matrixPrintAction(evt, "!version")
		version := fmt.Sprintf("gonalmatrix v%v.%v.%v, © 2021 BSDForen.de", VERSION_MAJOR, VERSION_MINOR, VERSION_PATCH)
		matrixClient.SendText(evt.RoomID, version)
	}
}

// ----

// Wrapper to start the syncer as goroutine. Must be
// called after the server connection was established!
func matrixSyncerWrapper(ch chan error) {
	err := matrixClient.Sync()
	ch <- err
}

// ----

// Authenticate against the matrix server.
func matrixAuthenticate(user string, passwd string) error {
	req := mautrix.ReqLogin{
		Type:             "m.login.password",
		Identifier:       mautrix.UserIdentifier{Type: mautrix.IdentifierTypeUser, User: user},
		Password:         passwd,
		StoreCredentials: true,
	}
	_, err := matrixClient.Login(&req)
	return err
}

// Deauthenticate from the matrix server.
func matrixDeauthenticate() error {
	_, err := matrixClient.Logout()
	if err != nil {
		return err
	}
	matrixClient.ClearCredentials()
	return err
}

// Connect to the given matrix home server.
func matrixConnect(homeserver string) error {
	client, err := mautrix.NewClient(homeserver, "", "")
	if err != nil {
		matrixClient = nil
	} else {
		matrixClient = client
	}
	return err
}

// Setup the event logger.
func matrixSetupLogger(handle io.Writer) {
	matrixLogger = log.New(handle, "", 0)
}

// Starts the syncer as goroutine.
// Returns an error channel to it.
func matrixStartSyncer() chan error {
	// Create syncer and register event handlers.
	syncer := matrixClient.Syncer.(*mautrix.DefaultSyncer)
	syncer.OnEventType(event.EventMessage, matrixHandleMessageEvent)

	// Add handler to ignore old events from
	// before the bot joined the rooms.
	var oei mautrix.OldEventIgnorer
	oei.Register(syncer)

	// And start the syncer.
	ch := make(chan error)
	go matrixSyncerWrapper(ch)
	return ch
}

// Stops the syncer.
func matrixStopSyncer() {
	matrixClient.StopSync()
}
