package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"

	"cloud.google.com/go/datastore"
	"google.golang.org/api/iterator"
)

type dsEntityStruct struct {
	InstrumentName string `json:"instrument_name" datastore:"instrument_name"`
	CaseID         string `json:"case_id" datastore:"case_id"`
	Disabled       bool   `json:"disabled" datastore:"disabled"`
}

func main() {

	projectID := os.Getenv("PROJECT_ID")
	uacsToEnable := os.Getenv("UACS_TO_ENABLE")
	uacsToDisable := os.Getenv("UACS_TO_DISABLE")

	uacEnableCount := 0
	uacEnableUpdatedCount := 0

	uacDisableCount := 0
	uacDisableUpdatedCount := 0

	enabled_uac_list := strings.Split(uacsToEnable, ",")
	disabled_uac_list := strings.Split(uacsToDisable, ",")

	ctx := context.Background()

	dsClient, err := datastore.NewClient(ctx, projectID)
	if err != nil {
		log.Fatal(err)
	}
	defer dsClient.Close()

	dsEnableQuery := datastore.NewQuery("uac").FilterField("Name/ID", "in", enabled_uac_list)
	dsDisableQuery := datastore.NewQuery("uac").FilterField("Name/ID", "in", disabled_uac_list)

	dsEnableIt := dsClient.Run(ctx, dsEnableQuery)
	dsEnableItCount := *dsEnableIt

	dsDisableIt := dsClient.Run(ctx, dsDisableQuery)
	dsDisableItCount := *dsDisableIt

	for {
		var entity dsEntityStruct
		_, err := dsEnableItCount.Next(&entity)
		if err == iterator.Done {
			break
		}
		if err != nil {
			log.Fatal(err)
		}
		uacEnableCount++
	}

	fmt.Println("Found", uacEnableCount, "UACs to enable")

	for {
		var entity dsEntityStruct
		_, err := dsDisableItCount.Next(&entity)
		if err == iterator.Done {
			break
		}
		if err != nil {
			log.Fatal(err)
		}
		uacDisableCount++
	}

	fmt.Println("Found", uacDisableCount, "UACs to disable")

	for {

		var entity dsEntityStruct

		key, err := dsEnableIt.Next(&entity)
		if err == iterator.Done {
			break
		}
		if err != nil {
			log.Fatal(err)
		}

		dsEntityKey := datastore.NameKey("uac", key.Name, nil)

		fmt.Println("Enabling UAC", key.Name)

		entity.Disabled = false

		if _, err := dsClient.Put(ctx, dsEntityKey, &entity); err != nil {
			log.Fatal(err)
		}

		fmt.Println("Enabled UAC", key.Name)

		uacEnableUpdatedCount++

	}

	fmt.Println("Enabled", uacEnableUpdatedCount, "UACs")

	for {

		var entity dsEntityStruct

		key, err := dsDisableIt.Next(&entity)
		if err == iterator.Done {
			break
		}
		if err != nil {
			log.Fatal(err)
		}

		dsEntityKey := datastore.NameKey("uac", key.Name, nil)

		fmt.Println("Disabling UAC", key.Name)

		entity.Disabled = true

		if _, err := dsClient.Put(ctx, dsEntityKey, &entity); err != nil {
			log.Fatal(err)
		}

		fmt.Println("Disabled UAC", key.Name)

		uacDisableUpdatedCount++

	}

	fmt.Println("Disabled", uacDisableUpdatedCount, "UACs")
}
