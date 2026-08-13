package uacgenerator

//go:generate mockery

import (
	"context"

	"cloud.google.com/go/datastore"
)

type UACServiceInterface interface {
	Generate(context.Context, string, []string) error
	GetAllUACs(context.Context, string) (UACs, error)
	GetAllUACsByCaseID(context.Context, string) (UACs, error)
	GetAllDisabledUACs(context.Context, string) (UACs, error)
	GetUACCount(context.Context, string) (int, error)
	GetUACInfo(context.Context, string) (*UACInfo, error)
	GetInstruments(context.Context) ([]string, error)
	ImportUACs(context.Context, []string) (int, error)
	AdminDelete(context.Context, string) error
	DisableUAC(context.Context, string) error
	EnableUAC(context.Context, string) error
}

type DatastoreInterface interface {
	Mutate(context.Context, ...*datastore.Mutation) ([]*datastore.Key, error)
	GetAll(context.Context, *datastore.Query, interface{}) ([]*datastore.Key, error)
	Count(context.Context, *datastore.Query) (int, error)
	Get(context.Context, *datastore.Key, interface{}) error
	DeleteMulti(context.Context, []*datastore.Key) error
	Close() error
}
