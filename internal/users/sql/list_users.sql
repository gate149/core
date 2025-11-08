SELECT *
from users
ORDER BY created_at DESC
LIMIT $1 OFFSET $2