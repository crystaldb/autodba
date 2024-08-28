# Metrics BFF

Exposes an api that can be configured to expose routes for executing arbitrary prometheus queries for time series results
See the configuration examples in config.json and config_all.json
config.json is the actual configuration the service will consume. 
config_all.json is an example that contains all the queries we might utilize

NOTE: There is a limit of 11 thousand samples per query. Meaning if you want data at a 5 second sample rate, you can only request a time span of 11000*5 seconds, about 15hours, in a single query. If you want data at a 30s sample rate, you can query up a 91 hour time range in a single query

## Testing

Run unit tests with: go test ./...
Run formatting with: go fmt ./...

## Setup Instructions
To run the project locally:

- In the base autodba project run : ./run.sh --db-url 'postgres://postgres:lTEP7OzeXQr77Ldu@mohammad-dashti-rds-1.cvirkksghnig.us-west-2.rds.amazonaws.com:5432/postgres?sslmode=require' --instance_id 1 --rds-instance mohammad-dashti-rds-1
- Use docker ps to check which port is being forwarded to 9090 (prometheus) and put that port in the config.json , eg "prometheus_server": "http://localhost:7001",
- In the bff folder root, run go build ./cmd/main.go
- in the bff folder root, run ./main, this will start up a server at the port in the config.json (default port 4000)
