#!/bin/sh

if [[ $# -ne 2 ]]; then
  echo "invalid."
  exit 1
fi

query="""
SELECT claims_logs.jwt FROM participants 
INNER JOIN claims_logs ON claims_logs.participant_id = participants.id 
WHERE participants.id = $2"""

echo $query | sqlite3 $1
