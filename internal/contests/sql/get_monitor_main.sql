WITH UserSolutions AS (
    SELECT cu.user_id,
        cp.problem_id,
        cp.position,
        s.state,
        s.created_at,
        ROW_NUMBER() OVER (
            PARTITION BY cu.user_id,
            cp.problem_id
            ORDER BY s.created_at
        ) AS attempt_number,
        MIN(
            CASE
                WHEN s.state = 200 THEN s.created_at
            END
        ) OVER (
            PARTITION BY cu.user_id,
            cp.problem_id
        ) AS first_success_time
    FROM contest_user cu
        JOIN contest_problem cp ON cu.contest_id = cp.contest_id
        LEFT JOIN solutions s ON cu.user_id = s.user_id
        AND cp.problem_id = s.problem_id
        AND cu.contest_id = s.contest_id
    WHERE cu.contest_id = $1
),
FailedAttempts AS (
    SELECT user_id,
        problem_id,
        position,
        COUNT(
            CASE
                WHEN state != 200
                AND state != 1
                AND (
                    first_success_time IS NULL
                    OR created_at < first_success_time
                ) THEN 1
            END
        ) AS failed_attempts,
        CASE
            WHEN BOOL_OR(state = 200) THEN 200
            ELSE MAX(state)
        END AS final_state
    FROM UserSolutions
    GROUP BY user_id,
        problem_id,
        position
)
SELECT user_id,
    problem_id,
    position,
    COALESCE(failed_attempts, 0) AS f_atts,
    final_state as state
FROM FailedAttempts
WHERE user_id IS NOT NULL
    AND problem_id IS NOT NULL
ORDER BY user_id,
    problem_id