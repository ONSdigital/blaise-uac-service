service: bus
runtime: go122

env_variables:
  DATASTORE_PROJECT: _DATASTORE_PROJECT
  BLAISE_BASE_URL: _BLAISE_BASE_URL
  SERVERPARK: _SERVERPARK
  GIN_MODE: release

vpc_access_connector:
  name: _VPC_CONNECTOR

basic_scaling:
  idle_timeout: 60s
  max_instances: 10

instance_class: B4

handlers:
- url: /.*
  script: auto
  secure: always
  redirect_http_response_code: 301
