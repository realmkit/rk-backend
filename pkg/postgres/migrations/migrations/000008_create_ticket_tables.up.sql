CREATE TABLE ticket_definitions (
    id uuid primary key,
    key text NOT NULL,
    name text NOT NULL,
    description text NOT NULL DEFAULT '',
    kind text NOT NULL,
    status text NOT NULL,
    default_team_group_id uuid,
    default_assignee_user_id uuid,
    submitter_can_close boolean NOT NULL DEFAULT true,
    submitter_can_reopen boolean NOT NULL DEFAULT true,
    allow_anonymous_submitter boolean NOT NULL DEFAULT false,
    requires_target_user boolean NOT NULL DEFAULT false,
    requires_punishment boolean NOT NULL DEFAULT false,
    requires_evidence boolean NOT NULL DEFAULT false,
    max_open_per_submitter integer NOT NULL DEFAULT 0,
    reopen_window_seconds bigint NOT NULL DEFAULT 0,
    sla_first_response_seconds bigint NOT NULL DEFAULT 0,
    sla_resolution_seconds bigint NOT NULL DEFAULT 0,
    metadata_schema_key text NOT NULL DEFAULT '',
    display_order integer NOT NULL DEFAULT 0,
    version bigint NOT NULL DEFAULT 1,
    created_at timestamptz NOT NULL,
    updated_at timestamptz NOT NULL,
    deleted_at timestamptz
);

CREATE TABLE tickets (
    id uuid primary key,
    definition_id uuid NOT NULL REFERENCES ticket_definitions(id),
    key text,
    title text NOT NULL,
    kind text NOT NULL,
    status text NOT NULL,
    priority text NOT NULL,
    submitter_user_id uuid,
    target_user_id uuid,
    punishment_id uuid,
    current_team_group_id uuid,
    assignee_user_id uuid,
    opened_at timestamptz NOT NULL,
    first_staff_response_at timestamptz,
    last_message_at timestamptz,
    last_message_author_user_id uuid,
    closed_at timestamptz,
    closed_by_user_id uuid,
    close_reason text NOT NULL DEFAULT '',
    resolution text NOT NULL DEFAULT '',
    escalation_level integer NOT NULL DEFAULT 0,
    sla_first_response_due_at timestamptz,
    sla_resolution_due_at timestamptz,
    message_count bigint NOT NULL DEFAULT 0,
    staff_message_count bigint NOT NULL DEFAULT 0,
    evidence_count bigint NOT NULL DEFAULT 0,
    version bigint NOT NULL DEFAULT 1,
    created_at timestamptz NOT NULL,
    updated_at timestamptz NOT NULL,
    deleted_at timestamptz
);

CREATE TABLE ticket_messages (
    id uuid primary key,
    ticket_id uuid NOT NULL REFERENCES tickets(id),
    author_user_id uuid,
    author_role text NOT NULL,
    visibility text NOT NULL,
    sequence bigint NOT NULL,
    content_format text NOT NULL,
    content_document_json jsonb NOT NULL,
    content_text text NOT NULL,
    content_checksum text NOT NULL DEFAULT '',
    edit_count bigint NOT NULL DEFAULT 0,
    version bigint NOT NULL DEFAULT 1,
    created_at timestamptz NOT NULL,
    updated_at timestamptz NOT NULL,
    deleted_at timestamptz
);

CREATE TABLE ticket_evidence (
    id uuid primary key,
    ticket_id uuid NOT NULL REFERENCES tickets(id),
    message_id uuid REFERENCES ticket_messages(id),
    asset_id uuid,
    external_url text NOT NULL DEFAULT '',
    label text NOT NULL DEFAULT '',
    description text NOT NULL DEFAULT '',
    visibility text NOT NULL,
    submitted_by_user_id uuid,
    created_at timestamptz NOT NULL,
    deleted_at timestamptz
);

CREATE TABLE ticket_actions (
    id uuid primary key,
    ticket_id uuid NOT NULL REFERENCES tickets(id),
    actor_user_id uuid,
    type text NOT NULL,
    status text NOT NULL,
    payload_json jsonb NOT NULL DEFAULT '{}',
    result_json jsonb NOT NULL DEFAULT '{}',
    idempotency_key text,
    created_at timestamptz NOT NULL,
    completed_at timestamptz,
    failed_at timestamptz,
    error text NOT NULL DEFAULT ''
);

CREATE UNIQUE INDEX idx_ticket_definitions_key_active ON ticket_definitions (key) WHERE deleted_at IS NULL;
CREATE INDEX idx_ticket_definitions_kind_status ON ticket_definitions (kind, status) WHERE deleted_at IS NULL;
CREATE INDEX idx_ticket_definitions_order ON ticket_definitions (display_order, id) WHERE deleted_at IS NULL;

CREATE UNIQUE INDEX idx_tickets_idempotency_key_active ON tickets (key) WHERE key IS NOT NULL AND deleted_at IS NULL;
CREATE INDEX idx_tickets_submitter_status ON tickets (submitter_user_id, status, updated_at DESC) WHERE deleted_at IS NULL;
CREATE INDEX idx_tickets_target_status ON tickets (target_user_id, status, updated_at DESC) WHERE deleted_at IS NULL;
CREATE INDEX idx_tickets_punishment ON tickets (punishment_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_tickets_team_queue ON tickets (current_team_group_id, status, priority, updated_at DESC) WHERE deleted_at IS NULL;
CREATE INDEX idx_tickets_assignee_queue ON tickets (assignee_user_id, status, updated_at DESC) WHERE deleted_at IS NULL;
CREATE INDEX idx_tickets_sla_due ON tickets (status, sla_first_response_due_at, sla_resolution_due_at) WHERE deleted_at IS NULL;

CREATE UNIQUE INDEX idx_ticket_messages_sequence_active ON ticket_messages (ticket_id, sequence) WHERE deleted_at IS NULL;
CREATE INDEX idx_ticket_messages_visibility ON ticket_messages (ticket_id, visibility, sequence) WHERE deleted_at IS NULL;
CREATE INDEX idx_ticket_messages_author ON ticket_messages (author_user_id, created_at DESC) WHERE deleted_at IS NULL;
CREATE INDEX idx_ticket_messages_text_search ON ticket_messages USING GIN (to_tsvector('simple', content_text));

CREATE INDEX idx_ticket_evidence_ticket ON ticket_evidence (ticket_id, created_at) WHERE deleted_at IS NULL;
CREATE INDEX idx_ticket_evidence_asset ON ticket_evidence (asset_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_ticket_evidence_message ON ticket_evidence (message_id) WHERE deleted_at IS NULL;

CREATE UNIQUE INDEX idx_ticket_actions_idempotency_active ON ticket_actions (idempotency_key) WHERE idempotency_key IS NOT NULL;
CREATE INDEX idx_ticket_actions_ticket ON ticket_actions (ticket_id, created_at);
CREATE INDEX idx_ticket_actions_type_status ON ticket_actions (type, status, created_at);
