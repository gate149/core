DELETE FROM contest_user
WHERE user_id = $1
    AND contest_id = $2