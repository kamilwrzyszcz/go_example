-- name: CreateArticle :one
INSERT INTO articles (
    author,
    headline,
    content
) VALUES (
    $1, $2, $3
) RETURNING *;

-- name: GetArticle :one
SELECT * FROM articles
WHERE id = $1 LIMIT 1;

-- name: UpdateArticle :one
UPDATE articles
SET
    headline = coalesce(sqlc.narg('headline'), headline),
    content = coalesce(sqlc.narg('content'), content),
    edited_at = NOW()
WHERE id = sqlc.arg('id')
RETURNING *;

-- name: ListArticles :many
SELECT * FROM articles
WHERE author = $1
ORDER BY id
LIMIT $2
OFFSET $3;