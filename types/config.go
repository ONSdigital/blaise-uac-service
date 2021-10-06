package types

type Config struct {
	Serverpark       string `default:"gusty"`
	DatastoreProject string `required:"true" split_words:"true"`
	BlaiseBaseUrl    string `required:"true" split_words:"true"`
	Port             string `default:"8082"`
}
