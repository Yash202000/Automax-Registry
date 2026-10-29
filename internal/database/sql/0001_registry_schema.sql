-- REG-01: one Registry database. All six tables are sub-concepts of the
-- Registry service, so they share storage and real foreign keys.
-- department_id is a plain reference to core's Department — no cross-database FK.

CREATE TABLE groups (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255) NOT NULL,
    parent_group_id UUID,
    department_id UUID NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    CONSTRAINT fk_groups_parent
        FOREIGN KEY (parent_group_id) REFERENCES groups (id) ON DELETE RESTRICT
);

CREATE INDEX idx_groups_parent_group_id ON groups (parent_group_id);
CREATE INDEX idx_groups_department_id ON groups (department_id);

CREATE TABLE things (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255) NOT NULL,
    type VARCHAR(50) NOT NULL,
    is_trackable BOOLEAN NOT NULL DEFAULT false,
    department_id UUID NOT NULL,
    group_id UUID,
    metadata JSONB NOT NULL DEFAULT '{}',
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    CONSTRAINT fk_things_group
        FOREIGN KEY (group_id) REFERENCES groups (id) ON DELETE RESTRICT
);

CREATE INDEX idx_things_type ON things (type);
CREATE INDEX idx_things_department_id ON things (department_id);
CREATE INDEX idx_things_group_id ON things (group_id);

CREATE TABLE channels (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255) NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE channel_acl (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    channel_id UUID NOT NULL,
    thing_id UUID NOT NULL,
    can_publish BOOLEAN NOT NULL DEFAULT false,
    can_subscribe BOOLEAN NOT NULL DEFAULT false,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    CONSTRAINT fk_channel_acl_channel
        FOREIGN KEY (channel_id) REFERENCES channels (id) ON DELETE CASCADE,
    CONSTRAINT fk_channel_acl_thing
        FOREIGN KEY (thing_id) REFERENCES things (id) ON DELETE CASCADE,
    CONSTRAINT uq_channel_acl_channel_thing UNIQUE (channel_id, thing_id)
);

CREATE INDEX idx_channel_acl_thing_id ON channel_acl (thing_id);

CREATE TABLE device_credentials (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    thing_id UUID NOT NULL,
    api_key VARCHAR(64) NOT NULL,
    api_secret_hash VARCHAR(255) NOT NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'active',
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    revoked_at TIMESTAMPTZ,
    CONSTRAINT fk_device_credentials_thing
        FOREIGN KEY (thing_id) REFERENCES things (id) ON DELETE RESTRICT,
    CONSTRAINT uq_device_credentials_api_key UNIQUE (api_key),
    CONSTRAINT ck_device_credentials_status CHECK (status IN ('active', 'revoked'))
);

CREATE INDEX idx_device_credentials_thing_id ON device_credentials (thing_id);

CREATE TABLE thing_external_refs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    thing_id UUID NOT NULL,
    external_system VARCHAR(50) NOT NULL,
    external_id VARCHAR(255) NOT NULL,
    synced_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    CONSTRAINT fk_thing_external_refs_thing
        FOREIGN KEY (thing_id) REFERENCES things (id) ON DELETE RESTRICT,
    CONSTRAINT uq_thing_external_refs_thing_system UNIQUE (thing_id, external_system)
);
