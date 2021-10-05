package main

import (
	"fmt"
	"strings"

	"maunium.net/go/mautrix"
	"maunium.net/go/mautrix/event"
)

// ----

// TODO: Should be done with a go routine and channel.
var matrixClient *mautrix.Client

// ----

func handleMaxtrixEvents(source mautrix.EventSource, event *event.Event) {
	bodyText := event.Content.Raw["body"].(string)
	roomID := event.RoomID

	// !ping -> Anwer with 'pong!'.
	if strings.HasPrefix(bodyText, "!ping") {
		matrixClient.SendText(roomID, "pong!")
	}
}

// ----

func connectMatrix(homeserver string, user string, passwd string) (*mautrix.Client, error) {
	fmt.Printf("Connecting to %v\n", homeserver)
	client, err := mautrix.NewClient(homeserver, "", "")
	if err != nil {
		return nil, err
	}

	fmt.Printf("Authenticating as %v\n", user)
	_, err = client.Login(&mautrix.ReqLogin{Type: "m.login.password", Identifier: mautrix.UserIdentifier{Type: mautrix.IdentifierTypeUser, User: user}, Password: passwd, StoreCredentials: true})
	if err != nil {
		return nil, err
	}

	return client, err
}

func startSyncer(client *mautrix.Client) (error) {
	syncer := client.Syncer.(*mautrix.DefaultSyncer)
	syncer.OnEventType(event.EventMessage, handleMaxtrixEvents)

	// TODO: Should be go routines and a channel!
	matrixClient = client
	err := client.Sync()

	return err
}
