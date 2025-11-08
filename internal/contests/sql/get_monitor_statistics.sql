SELECT cp.problem_id,
    COUNT(
        CASE
            WHEN s.state = 200 THEN 1
        END
    ) AS s_atts,
    COUNT(
        CASE
            WHEN s.state != 200
            AND s.state != 1 THEN 1
        END
    ) AS uns_atts,
    COUNT(*) AS t_atts,
    cp.position
FROM contest_problem cp
    LEFT JOIN solutions s ON cp.problem_id = s.problem_id
    AND cp.contest_id = s.contest_id
WHERE cp.contest_id = $1
GROUP BY (cp.problem_id, cp.position)
ORDER BY cp.problem_id