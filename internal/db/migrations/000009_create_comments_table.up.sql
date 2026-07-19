CREATE TABLE
    comments (
        id BIGSERIAL PRIMARY KEY,
        issue_id BIGINT NOT NULL REFERENCES issues (id) ON DELETE CASCADE,
        author_id BIGINT NOT NULL REFERENCES users (id) ON DELETE CASCADE,
        parent_comment_id BIGINT REFERENCES comments (id) ON DELETE CASCADE,
        content TEXT NOT NULL,
        is_edited BOOLEAN NOT NULL DEFAULT FALSE,
        created_at TIMESTAMPTZ NOT NULL DEFAULT NOW (),
        updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW ()
    );

CREATE INDEX idx_comments_issue_created_at ON comments (issue_id, created_at);

CREATE INDEX idx_parent_comment_id ON comments (parent_comment_id)
WHERE
    parent_comment_id IS NOT NULL;

CREATE INDEX idx_comments_author_id ON comments (author_id);