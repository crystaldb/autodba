#!/bin/bash

# Database file paths
OLD_DB_PATH="./bastatt-crystaldb-collector.db"  # Old database path
NEW_DB_PATH="./bastatt/crystaldb-collector.db"  # New database path

# Number of entries to insert
MAX_ENTRIES=200  # Set the maximum number of entries to migrate (can be changed)

# New prefix for the file storage path (inside Docker container)
NEW_FILE_PREFIX="/usr/local/autodba/share/collector_api_server/storage"

# Directory to check for the files (local directory)
LOCAL_DIR="./bastatt"

# Initialize the new database
initialize_new_db() {
    # Check if the new database file exists
    if [ ! -f "$NEW_DB_PATH" ]; then
        echo "New database not found. Creating a new database..."
    else
        echo "New database already exists. Skipping creation..."
    fi

    # Create the new database and tables
    sqlite3 "$NEW_DB_PATH" <<EOF
    -- Create the snapshots table
    CREATE TABLE IF NOT EXISTS snapshots (
        id INTEGER PRIMARY KEY AUTOINCREMENT,
        collected_at INTEGER,
        s3_location TEXT,
        system_id TEXT,
        system_scope TEXT,
        system_type TEXT
    );

    -- Create the compact_snapshots table (not used, commented out)
    -- CREATE TABLE IF NOT EXISTS compact_snapshots (
    --     id INTEGER PRIMARY KEY AUTOINCREMENT,
    --     collected_at INTEGER,
    --     s3_location TEXT,
    --     system_id TEXT,
    --     system_scope TEXT,
    --     system_type TEXT
    -- );
EOF
}

# Read from the old database and insert into the new one
migrate_data() {
    echo "Migrating data from old database to new database..."

    # Get the rows from the old snapshots table, ordered by 'collected_at' (oldest first)
    snapshots=$(sqlite3 "$OLD_DB_PATH" "SELECT collected_at, s3_location FROM snapshots ORDER BY collected_at LIMIT $MAX_ENTRIES;")
    
    if [ -z "$snapshots" ]; then
        echo "No snapshots found in the old database."
        return
    fi

    # Get the current timestamp
    current_timestamp=$(date +%s)

    # Initialize a counter for the offset
    offset=0

    # Initialize a counter for the number of entries inserted
    entries_inserted=0

    # Loop through each snapshot and insert it into the new database
    while IFS="|" read -r collected_at s3_location; do
        # Calculate the timestamp for the current row (10 seconds prior for each previous entry)
        new_collected_at=$((current_timestamp - offset))

        # Extract the last part of the s3_location (the file name)
        file_name=$(basename "$s3_location")

        # Construct the new s3_location with the updated prefix
        new_s3_location="$NEW_FILE_PREFIX/$file_name"

        # Check if the file exists in the local directory
        if [ ! -f "$LOCAL_DIR/$file_name" ]; then
            echo "File '$file_name' does not exist in '$LOCAL_DIR'. Skipping insertion."
            continue  # Skip this iteration and move to the next snapshot
        fi

        # Mock data for the new fields
        system_id="mock_system_id"
        system_scope="mock_system_scope"
        system_type="mock_system_type"

        # Insert the data into the new database
        sqlite3 "$NEW_DB_PATH" <<EOF
        INSERT INTO snapshots (collected_at, s3_location, system_id, system_scope, system_type)
        VALUES ($new_collected_at, '$new_s3_location', '$system_id', '$system_scope', '$system_type');
EOF

        # Uncommented: Remove the insertion into the compact_snapshots table

        # Increment the offset by 10 seconds for the next snapshot
        offset=$((offset + 10))

        # Increment the entries inserted counter
        entries_inserted=$((entries_inserted + 1))

        # Stop if we've inserted the maximum number of entries
        if [ "$entries_inserted" -ge "$MAX_ENTRIES" ]; then
            echo "Inserted $MAX_ENTRIES entries. Stopping migration."
            break
        fi
    done <<< "$snapshots"

    # Print how many entries were inserted
    echo "$entries_inserted entries successfully inserted into the new database."
}

# Initialize new database and migrate data
initialize_new_db
migrate_data

