SELECT cu.user_id,
    u.username,
    COUNT(DISTINCT s.problem_id) as solved_problems,
    0 as penalty
FROM contest_user cu
    LEFT JOIN solutions s ON cu.user_id = s.user_id
    AND cu.contest_id = s.contest_id
    AND s.state = 200
    LEFT JOIN users u ON cu.user_id = u.id
WHERE cu.contest_id = $1
GROUP BY (cu.user_id, u.username)