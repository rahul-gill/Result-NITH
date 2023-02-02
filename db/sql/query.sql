-- name: CreateStudent :one
INSERT INTO student(roll_number, name, fathers_name, batch, branch, latest_semester, cgpi)
VALUES (?, ?, ?, ?, ?, ?, ?)
RETURNING *;