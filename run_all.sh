#!/bin/bash

cd ~/pgAutoDBA && ./run.sh --db-url 'postgres://postgres:masterKey5@moe-autodba-experiments-2.cvirkksghnig.us-west-2.rds.amazonaws.com:5432/test?sslmode=require' --instance_id 1 2>&1 &

# instance 2: too little vacuum with recommendation to do more frequent vacuum.
cd ~/pgAutoDBA && ./run.sh --db-url 'postgres://postgres:masterKey5@moe-autodba-experiments-5.cvirkksghnig.us-west-2.rds.amazonaws.com:5432/test?sslmode=require' --instance_id 2 2>&1 &

# instance 3: too much vacuum with recommendation to do less frequent vacuum.
cd ~/pgAutoDBA && ./run.sh --db-url 'postgres://postgres:masterKey5@moe-autodba-experiments-4.cvirkksghnig.us-west-2.rds.amazonaws.com:5432/test?sslmode=require' --instance_id 3 2>&1 &

# instance-4 - johann-rds-5 - small DB, base instance (starting point)
# >instance-5 - johann-rds-6 - large DB, base instance (starting poing)
# >instance-7 - johann-rds-iops-6 - large DB, base instance, upgraded disk (recommended for large DB)
# instance 8 - johann-rds-medium-5 - small db, instance with extra DRAM (recommended for small DB)
./run.sh --db-url $(cat ~/rds-5.url) --instance_id 4 2>&1 &
./run.sh --db-url $(cat ~/rds-6.url) --instance_id 5 2>&1 &
./run.sh --db-url $(cat ~/rds-iops-6.url) --instance_id 7 2>&1 &
./run.sh --db-url $(cat ~/rds-medium-5.url) --instance_id 8 2>&1 &
