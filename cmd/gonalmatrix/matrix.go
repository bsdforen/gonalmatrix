package main

import (
	"context"
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
func matrixHandleMessageEvent(ctx context.Context, evt *event.Event) {
	message := evt.Content.AsMessage().Body

	// Mark the event as read.
	matrixClient.MarkRead(ctx, evt.RoomID, evt.ID)

	// Log everything that has a body.
	if len(message) != 0 {
		matrixLogMessageEvent(evt, message)
	}

	// !forget -> Delete a factoid.
	if strings.HasPrefix(strings.ToLower(message), "!forget") {
		split := strings.SplitN(message, " ", 2)

		if len(split) == 1 {
			// No argument -> error string.
			matrixPrintAction(evt, "!forget")
			matrixClient.SendText(ctx, evt.RoomID, "try: '!forget foo = bar' or '!forget foo'")
		} else if len(split) == 2 {
			// Argument -> Either remove part or everything.
			keyvalue := strings.SplitN(split[1], "=", 2)

			if len(keyvalue) == 1 {
				// User has given only the key -> remove everything.
				key := strings.Trim(keyvalue[0], " ")
				key = strings.ToLower(key)
				matrixPrintAction(evt, fmt.Sprintf("!forget %v", key))

				err := sqliteFactoidForget(key)
				if err != nil {
					matrixPrintError(evt, err)
				} else {
					matrixClient.SendText(ctx, evt.RoomID, fmt.Sprintf("forgot everything i knew about '%v'", key))
				}
			} else if len(keyvalue) == 2 {
				// User has given a key and a value -> remove part.
				key := strings.Trim(keyvalue[0], " ")
				key = strings.ToLower(key)
				value := strings.Trim(keyvalue[1], " ")
				matrixPrintAction(evt, fmt.Sprintf("!forget %v = %v", key, value))

				// The key and the value must not be empty.
				if len(key) == 0 || len(value) == 0 {
					matrixClient.SendText(ctx, evt.RoomID, "try: '!forget foo = bar' or '!forget foo'")
				} else {
					err := sqliteFactoidForgetValue(key, value)
					if err != nil {
						matrixPrintError(evt, err)
					} else {
						matrixClient.SendText(ctx, evt.RoomID, fmt.Sprintf("forgot %v = %v", key, value))
					}
				}
			}
		}
	}

	// !info -> Answer with factoid.
	if strings.HasPrefix(strings.ToLower(message), "!info") {
		split := strings.SplitN(message, " ", 2)

		if len(split) == 1 {
			// No argument -> random factoid.
			matrixPrintAction(evt, "!info")

			fact, err := sqliteFactoidGetRandom()
			if err != nil {
				matrixPrintError(evt, err)
			} else {
				matrixClient.SendText(ctx, evt.RoomID, fact)
			}
		} else if len(split) == 2 {
			// Argument -> all factoids with that key.
			key := strings.Trim(split[1], " ")
			key = strings.ToLower(key)
			matrixPrintAction(evt, fmt.Sprintf("!info %v", key))

			fact, err := sqliteFactoidGetForKey(key)
			if err != nil {
				matrixPrintError(evt, err)
			} else {
				matrixClient.SendText(ctx, evt.RoomID, fact)
			}
		}
	}

	// !learn -> Save a factoid.
	if strings.HasPrefix(strings.ToLower(message), "!learn") {
		split := strings.SplitN(message, " ", 2)

		if len(split) == 1 {
			// No argument -> error string.
			matrixPrintAction(evt, "!learn")
			matrixClient.SendText(ctx, evt.RoomID, "try: '!learn foo = bar'")
		} else if len(split) == 2 {
			// Argument -> save that factoid.
			keyvalue := strings.SplitN(split[1], "=", 2)

			if len(keyvalue) == 1 {
				// User has given a key but no value.
				key := strings.Trim(keyvalue[0], " ")
				key = strings.ToLower(key)
				matrixPrintAction(evt, fmt.Sprintf("!learn %v", key))
				matrixClient.SendText(ctx, evt.RoomID, "try: '!learn foo = bar'")
			} else if len(keyvalue) == 2 {
				// User has given a key and a value.
				key := strings.Trim(keyvalue[0], " ")
				key = strings.ToLower(key)
				value := strings.Trim(keyvalue[1], " ")
				matrixPrintAction(evt, fmt.Sprintf("!learn %v = %v", key, value))

				// The key and the value must not be empty.
				if len(key) == 0 || len(value) == 0 {
					matrixClient.SendText(ctx, evt.RoomID, "try: '!learn foo = bar'")
				} else {
					err := sqliteFactoidSet(key, value, evt.Sender.String(), evt.RoomID.String())
					if err != nil {
						matrixPrintError(evt, err)
					} else {
						matrixClient.SendText(ctx, evt.RoomID, fmt.Sprintf("Okay, learned %v = %v", key, value))
					}
				}
			}
		}
	}

	// !ping -> Anwer with 'pong!'.
	if strings.HasPrefix(strings.ToLower(message), "!ping") {
		matrixPrintAction(evt, "!ping")
		matrixClient.SendText(ctx, evt.RoomID, "pong!")
	}

	// !version -> Answer with the version numer.
	if strings.HasPrefix(strings.ToLower(message), "!version") {
		matrixPrintAction(evt, "!version")
		version := fmt.Sprintf("gonalmatrix v%v.%v.%v, © 2021, 2023 - 2024 BSDForen.de", VERSION_MAJOR, VERSION_MINOR, VERSION_PATCH)
		matrixClient.SendText(ctx, evt.RoomID, version)
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
func matrixAuthenticate(ctx context.Context, user string, passwd string) error {
	req := mautrix.ReqLogin{
		Type:             "m.login.password",
		Identifier:       mautrix.UserIdentifier{Type: mautrix.IdentifierTypeUser, User: user},
		Password:         passwd,
		StoreCredentials: true,
	}
	_, err := matrixClient.Login(ctx, &req)
	return err
}

// Deauthenticate from the matrix server.
func matrixDeauthenticate(ctx context.Context) error {
	_, err := matrixClient.Logout(ctx)
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
func matrixStartSyncer(ctx context.Context) chan error {
	// Create syncer and register event handlers.
	syncer := matrixClient.Syncer.(*mautrix.DefaultSyncer)
	syncer.OnEventType(event.EventMessage, matrixHandleMessageEvent)

	// Add handler to ignore old events from
	// before the bot joined the rooms.
	syncer.OnSync(matrixClient.DontProcessOldEvents)

	// Start the syncer.
	ch := make(chan error)
	go matrixSyncerWrapper(ch)

	// Set our presence to online.
	matrixClient.SetPresence(ctx, event.PresenceOnline)
	return ch
}

// Stops the syncer.
func matrixStopSyncer() {
	matrixClient.StopSync()
}
