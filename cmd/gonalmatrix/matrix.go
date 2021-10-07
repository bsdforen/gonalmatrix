package main

import (
	"strings"

	"maunium.net/go/mautrix"
	"maunium.net/go/mautrix/event"
)

// ----

// Client instance, representing the server connection.
var matrixClient *mautrix.Client

// ----

func matrixHandleMessageEvent(source mautrix.EventSource, evt *event.Event) {
	content := evt.Content.AsMessage()

	// !ping -> Anwer with 'pong!'.
	if strings.HasPrefix(content.Body, "!ping") {
		matrixClient.SendText(evt.RoomID, "pong!")
	}
}

// ----

func matrixSyncerWrapper(ch chan error) {
	err := matrixClient.Sync()
	ch <- err
}

// ----

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

func matrixConnect(homeserver string) error {
	client, err := mautrix.NewClient(homeserver, "", "")
	if err != nil {
		matrixClient = nil
	} else {
		matrixClient = client
	}
	return err
}

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
