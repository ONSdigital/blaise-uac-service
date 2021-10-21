module github.com/ONSDigital/blaise-uac-service

go 1.15

require (
	cloud.google.com/go/datastore v1.1.0
	github.com/GoogleCloudPlatform/opentelemetry-operations-go/exporter/trace v1.0.0
	github.com/gin-gonic/gin v1.7.4
	github.com/golang/mock v1.6.0 // indirect
	github.com/googleapis/gax-go v1.0.3 // indirect
	github.com/jarcoal/httpmock v1.0.8
	github.com/kelseyhightower/envconfig v1.4.0
	github.com/onsi/ginkgo v1.16.4
	github.com/onsi/gomega v1.14.0
	github.com/stretchr/testify v1.7.0
	github.com/vektra/mockery/v2 v2.9.4 // indirect
	github.com/zenthangplus/goccm v0.0.0-20200608171100-39e9e08b694a
	go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin v0.25.0
	go.opentelemetry.io/otel v1.0.1
	go.opentelemetry.io/otel/sdk v1.0.1
	golang.org/x/sync v0.0.0-20210220032951-036812b2e83c
	google.golang.org/genproto v0.0.0-20210722135532-667f2b7c528f
	google.golang.org/grpc v1.39.0
)
