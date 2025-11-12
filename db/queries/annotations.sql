-- name: CreateAnnotation :one
INSERT INTO annotations (image_id, username, stage_index, option_value)
VALUES (?, ?, ?, ?)
ON CONFLICT(image_id, username, stage_index)
DO UPDATE SET
  option_value = excluded.option_value,
  annotated_at = CURRENT_TIMESTAMP
RETURNING *;

-- name: GetAnnotation :one
SELECT * FROM annotations
WHERE image_id = ? AND username = ? AND stage_index = ?;

-- name: GetAnnotationsForImage :many
SELECT * FROM annotations
WHERE image_id = ?
ORDER BY stage_index ASC;

-- name: GetAnnotationsByUser :many
SELECT a.*, i.path, i.original_filename
FROM annotations a
JOIN images i ON a.image_id = i.id
WHERE a.username = ?
ORDER BY a.annotated_at DESC
LIMIT ? OFFSET ?;

-- name: GetAnnotationsByImageAndUser :many
SELECT * FROM annotations
WHERE image_id = ? AND username = ?
ORDER BY stage_index ASC;

-- name: CountAnnotationsByUser :one
SELECT COUNT(*) FROM annotations
WHERE username = ?;

-- name: ListPendingImagesForUserAndStage :many
WITH annotated_images AS (
  SELECT image_id FROM annotations WHERE username = ? AND stage_index = ?
)
SELECT i.*
FROM images i
LEFT JOIN annotated_images ai ON i.id = ai.image_id
WHERE i.is_finished = FALSE AND ai.image_id IS NULL
ORDER BY i.completed_stages ASC, i.id ASC
LIMIT ?;

-- name: CheckAnnotationExists :one
SELECT EXISTS (
    SELECT 1
    FROM annotations
    WHERE image_id = ? AND username = ? AND stage_index = ?
);

-- name: DeleteAnnotation :exec
DELETE FROM annotations
WHERE id = ?;

-- name: DeleteAnnotationsForImage :exec
DELETE FROM annotations
WHERE image_id = ?;

-- name: GetAnnotationStats :one
SELECT
  COUNT(DISTINCT image_id) as annotated_images,
  COUNT(*) as total_annotations,
  COUNT(DISTINCT username) as total_users
FROM annotations;

-- name: CountPendingImagesForUserAndStage :one
WITH annotated_images AS (
  SELECT image_id FROM annotations WHERE username = ? AND stage_index = ?
)
SELECT COUNT(*)
FROM images i
LEFT JOIN annotated_images ai ON i.id = ai.image_id
WHERE i.is_finished = FALSE AND ai.image_id IS NULL;

-- name: GetImageIDsWithAnnotation :many
SELECT DISTINCT image_id
FROM annotations
WHERE stage_index = ? AND option_value = ?;

-- name: CountImagesWithoutAnnotationForStage :one
WITH annotated_images AS (
  SELECT DISTINCT image_id FROM annotations WHERE stage_index = ?
)
SELECT COUNT(*)
FROM images i
LEFT JOIN annotated_images ai ON i.id = ai.image_id
WHERE i.is_finished = FALSE AND ai.image_id IS NULL;
