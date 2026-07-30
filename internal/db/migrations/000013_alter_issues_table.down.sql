ALTER TABLE issues
DROP COLUMN due_date;

CREATE TYPE activity_type_new AS ENUM (
    'ISSUE_CREATED',
    'ISSUE_DETAILS_UPDATED',
    'ISSUE_STATUS_CHANGED',
    'ISSUE_PRIORITY_CHANGED',
    'ISSUE_ASSIGNEE_CHANGED',
    'ISSUE_DELETED',
    'ISSUE_RESTORED',
    'COMMENT_CREATED',
    'COMMENT_UPDATED',
    'COMMENT_DELETED'
);

ALTER TABLE issue_activities
ALTER COLUMN activity_type
TYPE activity_type_new
USING (
    CASE
        WHEN activity_type::text = 'DUE_DATE_CHANGED'
            THEN 'ISSUE_DETAILS_UPDATED'
        ELSE activity_type::text
    END
)::activity_type_new;

DROP TYPE activity_type;

ALTER TYPE activity_type_new
RENAME TO activity_type;