package main

import (
	"fmt"

	"maunium.net/go/mautrix"
)

// ----

func connectMatrix(homeserver string, user string, passwd string) (*mautrix.Client, error) {
	fmt.Printf("Connecting to %v\n", homeserver)
	client, err := mautrix.NewClient(homeserver, "", "")
	if err != nil {
		return nil, err
	}

	fmt.Printf("Authenticating as %v\n", user)
	_, err = client.Login(&mautrix.ReqLogin{Type: "m.login.password", Identifier: mautrix.UserIdentifier{Type: mautrix.IdentifierTypeUser, User: user}, Password: passwd})
	if err != nil {
		return nil, err
	}

	return client, err
}
