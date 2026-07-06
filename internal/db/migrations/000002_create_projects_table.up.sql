CREATE TABLE
    projects (
        id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
        owner_id BIGINT NOT NULL REFERENCES users (id) ON DELETE CASCADE,
        name TEXT NOT NULL,
        description TEXT NOT NULL,
        created_at TIMESTAMPTZ NOT NULL DEFAULT NOW (),
        updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW (),
        CONSTRAINT projects_owner_id_name_key UNIQUE (owner_id, name)
    );