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

	uacsToDisable := []string{}
	if envValue := os.Getenv("UACS_TO_DISABLE"); envValue != "" {
		uacsToDisable = strings.Split(envValue, ",")

		for i := range uacsToDisable {
			uacsToDisable[i] = strings.TrimSpace(uacsToDisable[i])
		}
	}

	ctx := context.Background()
	dsClient, err := datastore.NewClient(ctx, projectID)
	if err != nil {
		log.Fatal(err)
	}
	defer func(dsClient *datastore.Client) {
		err := dsClient.Close()
		if err != nil {
			log.Fatal(err)
		}
	}(dsClient)

	for _, keyName := range uacsToDisable {

		key := datastore.NameKey("uac", keyName, nil)

		var entity dsEntityStruct
		err := dsClient.Get(ctx, key, &entity)
		if err != nil {
			log.Fatal(err)
		}

		fmt.Println("Disabling UAC", keyName)
		entity.Disabled = true

		if _, err := dsClient.Put(ctx, key, &entity); err != nil {
			log.Fatal(err)
		}
	}

}
