package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"

	"cloud.google.com/go/datastore"
)

type dsEntityStruct struct {
	InstrumentName string `json:"instrument_name" datastore:"instrument_name"`
	CaseID         string `json:"case_id" datastore:"case_id"`
	Disabled       bool   `json:"disabled" datastore:"disabled"`
}

func main() {

	projectID := os.Getenv("PROJECT_ID")
	uacsToEnable := strings.Split(os.Getenv("UACS_TO_ENABLE"), ",")
	uacsToDisable := strings.Split(os.Getenv("UACS_TO_DISABLE"), ",")

	ctx := context.Background()
	dsClient, err := datastore.NewClient(ctx, projectID)
	if err != nil {
		log.Fatal(err)
	}
	defer func(dsClient *datastore.Client) {
		err := dsClient.Close()
		if err != nil {

		}
	}(dsClient)

	changeDisabledState(uacsToEnable, dsClient, ctx, false)
	changeDisabledState(uacsToDisable, dsClient, ctx, true)
}

func changeDisabledState(uacs []string, dsClient *datastore.Client, ctx context.Context, disabled bool) {
	for _, keyName := range uacs {

		key := datastore.NameKey("uac", keyName, nil)

		var entity dsEntityStruct
		err := dsClient.Get(ctx, key, &entity)
		if err != nil {
			return
		}

		message := func(disabled bool) string {
			if disabled {
				return "Disabling UAC"
			}
			return "Enabling UAC"
		}(disabled)

		fmt.Println(message, keyName)
		entity.Disabled = disabled

		if _, err := dsClient.Put(ctx, key, &entity); err != nil {
			log.Fatal(err)
		}
	}
}
