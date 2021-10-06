package types

import (
	"context"

	"cloud.google.com/go/datastore"
)

//Generate mocks by running "go generate ./..."
//go:generate mockery --name Datastore
type Datastore interface {
	Mutate(context.Context, ...*datastore.Mutation) ([]*datastore.Key, error)
	GetAll(context.Context, *datastore.Query, interface{}) ([]*datastore.Key, error)
	Count(context.Context, *datastore.Query) (int, error)
	Get(context.Context, *datastore.Key, interface{}) error
	DeleteMulti(context.Context, []*datastore.Key) error
	Close() error
}

type UacChunks struct {
	UAC1 string `json:"uac1"`
	UAC2 string `json:"uac2"`
	UAC3 string `json:"uac3"`
	UAC4 string `json:"uac4,omitempty"`
}
