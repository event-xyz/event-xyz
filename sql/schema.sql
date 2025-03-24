.mode box

.print "checkpoints"
PRAGMA table_info(checkpoints);
.print "participants"
PRAGMA table_info(participants);
.print "teams"
PRAGMA table_info(teams);
.print "db_participants"
PRAGMA table_info(db_participants);
.print "db_authorised_users"
PRAGMA table_info(db_authoriesed_users);

.print "claims_logs"
PRAGMA table_info(claims_logs);

.print "\n"

.schema

-- select name,phone,checkin,entry_time,checkout,exit_time,breakfast,dinner,snacks from db_participants;

-- add more for tests :)
