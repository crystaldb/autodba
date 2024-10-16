# Integration tests

## Running the Tests

To run the tests, use the following command structure, replacing the placeholders with your actual database details:
The dbconfig json is a map of postgres versions to details for a db of that version. This corelation will be validated in the tests

```bash

go test ./... -run TestAPISuite -v -timeout 20m -args -dbconfig='{
  "16": {
    "description": "PostgreSQL 16 instance",
    "host": "your-db-host",
    "name": "your-db-name",
    "username": "your-db-username",
    "password": "your-db-password",
    "port": "5432",
    "aws_rds_instance": "your-instance-id",
    "aws_region": "your-aws-region",
    "aws_access_key": "your-aws-access-key",
    "aws_secret": "your-aws-secret-key",
    "system_scope": "us-west-2/abcdefghijk",
    "system_type": "amazon_rds"
  }
}'
```
