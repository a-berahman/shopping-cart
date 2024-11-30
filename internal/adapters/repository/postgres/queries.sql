-- name: CreateItem :one
INSERT INTO items (
    name, 
    quantity, 
    status
) VALUES (
    $1, $2, $3
) RETURNING *;

-- name: ListItems :many
SELECT * FROM items 
ORDER BY created_at DESC;

-- name: UpdateItemReservation :one
UPDATE items 
SET reservation_id = $2,
    status = $3,
    updated_at = NOW()
WHERE id = $1
RETURNING *;

-- name: GetItem :one
SELECT * FROM items 
WHERE id = $1;

-- name: UpdateItemStatus :one
UPDATE items 
SET status = $2,
    updated_at = NOW()
WHERE id = $1
RETURNING *;