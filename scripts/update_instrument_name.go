package main

import (
	"cloud.google.com/go/datastore"
	"context"
	"fmt"
	"google.golang.org/api/iterator"
	"log"
	"os"
)

type dsEntityStruct struct {
	InstrumentName string `json:"instrument_name" datastore:"instrument_name"`
	CaseID         string `json:"case_id" datastore:"case_id"`
	Disabled       bool   `json:"disabled" datastore:"disabled"`
}

func main() {

	projectID := os.Getenv("PROJECT_ID")
	oldInstrumentName := os.Getenv("OLD_INSTRUMENT_NAME")
	newInstrumentName := os.Getenv("NEW_INSTRUMENT_NAME")
	uacCount := 0
	uacUpdatedCount := 0

	ctx := context.Background()

	dsClient, err := datastore.NewClient(ctx, projectID)
	if err != nil {
		log.Fatal(err)
	}
	defer dsClient.Close()

	dsQuery := datastore.NewQuery("uac").FilterField("instrument_name", "=", oldInstrumentName)

	dsIt := dsClient.Run(ctx, dsQuery)
	dsItCount := *dsIt

	for {
		var entity dsEntityStruct
		_, err := dsItCount.Next(&entity)
		if err == iterator.Done {
			break
		}
		if err != nil {
			log.Fatal(err)
		}
		uacCount++
	}

	fmt.Println("Found", uacCount, "UACs for instrument name", oldInstrumentName)

	for {

		var entity dsEntityStruct

		key, err := dsIt.Next(&entity)
		if err == iterator.Done {
			break
		}
		if err != nil {
			log.Fatal(err)
		}

		dsEntityKey := datastore.NameKey("uac", key.Name, nil)

		fmt.Println("Updating UAC", key.Name, "from instrument name", oldInstrumentName, "to instrument name", newInstrumentName)

		entity.InstrumentName = newInstrumentName

		if _, err := dsClient.Put(ctx, dsEntityKey, &entity); err != nil {
			log.Fatal(err)
		}

		fmt.Println("Updated UAC", key.Name, "from instrument name", oldInstrumentName, "to instrument name", newInstrumentName)

		uacUpdatedCount++

	}

	fmt.Println("Updated", uacUpdatedCount, "UACs for instrument name", oldInstrumentName)
}
