CREATE TABLE
    issue_attachments (
        id BIGSERIAL PRIMARY KEY,
        issue_id BIGINT NOT NULL REFERENCES issues (id) ON DELETE CASCADE,
        comment_id BIGINT REFERENCES comments (id) ON DELETE CASCADE,
        uploaded_by BIGINT NOT NULL REFERENCES users (id) ON DELETE RESTRICT,
        original_name TEXT NOT NULL,
        storage_key TEXT NOT NULL,
        resource_type VARCHAR(20) NOT NULL,
        mime_type VARCHAR(100) NOT NULL,
        file_size BIGINT NOT NULL CHECK (file_size > 0),
        created_at TIMESTAMPTZ NOT NULL DEFAULT NOW (),
        deleted_at TIMESTAMPTZ
    );

CREATE INDEX idx_issue_attachments_issue 
ON issue_attachments (issue_id)
WHERE deleted_at IS NULL;

CREATE INDEX idx_issue_attachments_comment 
ON issue_attachments (comment_id)
WHERE comment_id IS NOT NULL AND deleted_at IS NULL;

CREATE UNIQUE INDEX idx_issue_attachments_unique_storage_key 
ON issue_attachments (storage_key)
WHERE deleted_at IS NULL;

CREATE INDEX idx_issue_attachments_uploader 
ON issue_attachments (uploaded_by);