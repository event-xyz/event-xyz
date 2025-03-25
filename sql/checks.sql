.mode box

-- check participants checkedin
select participants.id, participants.name, checkpoints.checkin, checkpoints.entry_time, checkpoints.checkout, 
checkpoints.exit_time, checkpoints.snacks, checkpoints.breakfast, checkpoints.dinner 
FROM db_participants 
INNER JOIN participants ON participants.id = db_participants.participant_id 
INNER JOIN checkpoints ON db_participants.checkpoints_id = checkpoints.id;
