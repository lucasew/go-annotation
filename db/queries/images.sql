-- name: CreateImage :one
INSERT INTO images (path, original_filename)
VALUES (?, ?)
RETURNING *;

-- name: GetImage :one
SELECT * FROM images
WHERE id = ?;

-- name: GetImageByPath :one
SELECT * FROM images
WHERE path = ?;

-- name: ListImages :many
SELECT * FROM images
ORDER BY id;

-- name: ListImagesNotFinished :many
SELECT * FROM images
WHERE is_finished = FALSE
ORDER BY completed_stages ASC, id ASC
LIMIT ?;

-- name: UpdateImageCompletionStatus :exec
UPDATE images
SET completed_stages = ?,
    is_finished = ?
WHERE id = ?;

-- name: CountImages :one
SELECT COUNT(*) FROM images;

-- name: CountPendingImages :one
SELECT COUNT(*) FROM images
WHERE is_finished = FALSE;

-- name: DeleteImage :exec
DELETE FROM images
WHERE id = ?;
