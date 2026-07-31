CREATE TABLE
    issue_labels (
        issue_id BIGINT NOT NULL REFERENCES issues (id) ON DELETE CASCADE,
        label_id BIGINT NOT NULL REFERENCES labels (id) ON DELETE CASCADE,
        created_at TIMESTAMPTZ NOT NULL DEFAULT NOW (),
        PRIMARY KEY (issue_id, label_id)
    );

CREATE INDEX idx_issue_labels_issues 
ON issue_labels (issue_id);

CREATE INDEX idx_issue_labels_label_id
ON issue_labels (label_id);