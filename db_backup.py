# Python script to backup the SQLite DB from host to backup machine using SCP
# Setup SSH keys before running the script

import subprocess
import time
import sys

if len(sys.argv) < 4:
    print("Invalid arguments!")
    print(
        "Usage: python3 db_backup.py <db_origin_absolute_path_with_host> <db_destination_absolute_path_with destination_file_name>(as the current time is suffixed to backups) <backup_interval(seconds)>"
    )
    sys.exit(1)

DB_ORIGIN_PATH = sys.argv[1]
DB_DESTINATION_PATH = sys.argv[2]
try:
    BACKUP_INTERVAL = int(sys.argv[3])
except ValueError:
    print("Invalid backup interval!")
    sys.exit(1)

while True:
    current_time = "__" + time.ctime().replace(" ", "_")
    print(
        subprocess.run(
            ["scp", DB_ORIGIN_PATH, DB_DESTINATION_PATH + current_time],
            capture_output=True,
        )
    )
    time.sleep(BACKUP_INTERVAL)
