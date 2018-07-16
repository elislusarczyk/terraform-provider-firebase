package firebase

import (
	"context"
	"log"

	firebase "firebase.google.com/go"

	"firebase.google.com/go/auth"
	"firebase.google.com/go/db"
	"firebase.google.com/go/messaging"
	"firebase.google.com/go/storage"

	"google.golang.org/api/option"
)

type Config struct {
	ServiceAccountKey string
}

type Client struct {
	App       firebase.App
	Auth      *auth.Client
	DB        *db.Client
	Messaging *messaging.Client
	Storage   *storage.Client
}

// Client configures and returns a fully initialized firebase app client
func (c Config) Client() (interface{}, error) {
	var err error
	var client Client
	ctx := context.Background()

	log.Printf("[INFO] Using service account key file `%s`\n", c.ServiceAccountKey)

	opt := option.WithCredentialsFile(c.ServiceAccountKey)

	log.Println("[INFO] Create new firebase app client")

	app, err := firebase.NewApp(ctx, nil, opt)
	if err != nil {
		return nil, err
	}

	log.Println("[INFO] Getting auth client")

	// Get an auth client from the firebase.App
	client.Auth, err = app.Auth(ctx)
	if err != nil {
		return nil, err
	}

	// log.Println("[INFO] Getting database client")
	//
	// // Get a database client from the firebase.App
	// client.DB, err = app.Database(ctx)
	// if err != nil {
	// 	return nil, err
	// }

	log.Println("[INFO] Getting messaging client")

	// Get a messaging client from the firebase.App
	client.Messaging, err = app.Messaging(ctx)
	if err != nil {
		return nil, err
	}

	log.Println("[INFO] Getting storage client")

	// Get a storage client from the firebase.App
	client.Storage, err = app.Storage(ctx)
	if err != nil {
		return nil, err
	}

	return client, nil
}
