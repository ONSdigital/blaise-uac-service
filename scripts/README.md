# Yay some scripts...

Just some useful support scripts for running locally...

## update_instrument_name

Updates all instances of an instrument_name! #ronseal

Obviously be **VERY CAREFUL** if running in prod!

Login to the GCP CLI:
```
gcloud auth login
```

Point the GCP CLI to the appropriate project:
```
gcloud config set project ons-blaise-v2-dev-<sandbox>
```

Get a service account JSON key:
```
gcloud iam service-accounts keys create keys.json --iam-account ons-blaise-v2-dev-<sandbox>@appspot.gserviceaccount.com
```

Set some local env vars:

Unix:
```
export GOOGLE_APPLICATION_CREDENTIALS=keys.json
export PROJECT_ID=ons-blaise-v2-dev-<sandbox>
export OLD_INSTRUMENT_NAME=lms2212_rr1
export NEW_INSTRUMENT_NAME=lms2212_rr5
```

Windows:
```
set GOOGLE_APPLICATION_CREDENTIALS=keys.json
set PROJECT_ID=ons-blaise-v2-dev-<sandbox>
set OLD_INSTRUMENT_NAME=lms2212_rr1
set NEW_INSTRUMENT_NAME=lms2212_rr5
```

Run da ting:
```
go run update_instrument_name.go
```