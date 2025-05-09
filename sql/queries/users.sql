-- name: CreateUser :one
INSERT INTO users (email,hashed_password,created_at,updated_at) VALUES($1,$2,NOW(),NOW()) RETURNING *;

-- name: DeleteUser :exec
DELETE FROM users; 

-- name: CreateChirp :one
INSERT INTO chirps (body,user_id,created_at,updated_at) VALUES($1,$2,NOW(),NOW()) RETURNING *;

-- name: GetAllChirps :many
SELECT * FROM chirps ORDER by created_at;

-- name: GetOneUser :one
SELECT * FROM chirps WHERE id=$1 ;

-- name: GetOneUserByEmail :one
SELECT * FROM users WHERE LOWER(email)=LOWER($1) ;
