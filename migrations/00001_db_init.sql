-- +goose Up
-- +goose StatementBegin
BEGIN;

CREATE EXTENSION IF NOT EXISTS pgcrypto;

CREATE TABLE IF NOT EXISTS imports (
    import_id     uuid PRIMARY KEY,
    uploaded_by   uuid NULL,
    file_name     text NOT NULL,
    file_sha256   text NOT NULL,
    status        text NOT NULL,
    total_rows    int  NOT NULL DEFAULT 0,
    inserted_rows int  NOT NULL DEFAULT 0,
    skipped_rows  int  NOT NULL DEFAULT 0,
    created_at    timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS ix_imports_created_at ON imports(created_at DESC);
CREATE INDEX IF NOT EXISTS ix_imports_status ON imports(status);

CREATE TABLE IF NOT EXISTS candidates (
    candidate_id uuid PRIMARY KEY,
    first_name   text NOT NULL,
    last_name    text NOT NULL,
    birth_year   int  NULL,
    citizenship  text NULL,
    languages    text NULL,
    created_at   timestamptz NOT NULL DEFAULT now(),
    updated_at   timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS ix_candidates_name ON candidates(last_name, first_name);
CREATE INDEX IF NOT EXISTS ix_candidates_updated_at ON candidates(updated_at DESC);

CREATE TABLE IF NOT EXISTS candidate_contacts (
    contact_id    uuid PRIMARY KEY,
    candidate_id  uuid NOT NULL,
    type          text NOT NULL,
    value         text NOT NULL,
    is_primary    bool NOT NULL DEFAULT true,
    normalized    text NOT NULL,
    created_at    timestamptz NOT NULL DEFAULT now()
);

CREATE UNIQUE INDEX IF NOT EXISTS ux_candidate_contacts_type_norm
    ON candidate_contacts(type, normalized);

CREATE INDEX IF NOT EXISTS ix_candidate_contacts_candidate
    ON candidate_contacts(candidate_id);

CREATE INDEX IF NOT EXISTS ix_candidate_contacts_candidate_primary
    ON candidate_contacts(candidate_id)
    WHERE is_primary = true;

CREATE TABLE IF NOT EXISTS applications (
    application_id   uuid PRIMARY KEY,
    candidate_id     uuid NOT NULL,
    import_id        uuid NOT NULL,
    applied_at       timestamptz NOT NULL,
    resume_url       text NULL,
    priority1        text NULL,
    priority2        text NULL,
    course           text NULL,
    specialty        text NULL,
    specialty_other  text NULL,
    schedule         text NULL,
    city             text NULL,
    city_other       text NULL,
    university       text NULL,
    university_other text NULL,
    source           text NULL,
    status           text NOT NULL,
    status_reason    text NULL,
    external_key     text NOT NULL,
    raw_row          jsonb NOT NULL,
    created_at       timestamptz NOT NULL DEFAULT now(),
    updated_at       timestamptz NOT NULL DEFAULT now()
);

CREATE UNIQUE INDEX IF NOT EXISTS ux_applications_external_key
    ON applications(external_key);

CREATE INDEX IF NOT EXISTS ix_applications_applied_at
    ON applications(applied_at DESC);

CREATE INDEX IF NOT EXISTS ix_applications_status
    ON applications(status);

CREATE INDEX IF NOT EXISTS ix_applications_candidate
    ON applications(candidate_id);

CREATE INDEX IF NOT EXISTS ix_applications_import
    ON applications(import_id);

CREATE INDEX IF NOT EXISTS ix_applications_status_applied_at
    ON applications(status, applied_at DESC);

CREATE INDEX IF NOT EXISTS gin_applications_raw_row
    ON applications USING gin (raw_row);

CREATE TABLE IF NOT EXISTS application_notes (
    note_id         uuid PRIMARY KEY,
    application_id  uuid NOT NULL,
    author_id       uuid NULL,
    note            text NOT NULL,
    created_at      timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS ix_application_notes_application
    ON application_notes(application_id, created_at DESC);

CREATE TABLE IF NOT EXISTS message_templates (
    template_id uuid PRIMARY KEY,
    code        text NOT NULL,
    channel     text NOT NULL,
    subject     text NOT NULL,
    body        text NOT NULL,
    is_active   bool NOT NULL DEFAULT true,
    created_at  timestamptz NOT NULL DEFAULT now(),
    updated_at  timestamptz NOT NULL DEFAULT now()
);

CREATE UNIQUE INDEX IF NOT EXISTS ux_message_templates_code
    ON message_templates(code);

CREATE INDEX IF NOT EXISTS ix_message_templates_active
    ON message_templates(is_active)
    WHERE is_active = true;

CREATE TABLE IF NOT EXISTS email_outbox (
    email_id             uuid PRIMARY KEY,
    application_id       uuid NOT NULL,
    to_email             text NOT NULL,
    template_id          uuid NOT NULL,
    render_vars          jsonb NOT NULL DEFAULT '{}'::jsonb,
    status               text NOT NULL,
    attempt              int  NOT NULL DEFAULT 0,
    next_retry_at        timestamptz NULL,
    locked_until         timestamptz NULL,
    provider_message_id  text NULL,
    last_error           text NULL,
    created_at           timestamptz NOT NULL DEFAULT now(),
    updated_at           timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS ix_email_outbox_pick
    ON email_outbox(status, next_retry_at, created_at);

CREATE INDEX IF NOT EXISTS ix_email_outbox_application
    ON email_outbox(application_id);

CREATE INDEX IF NOT EXISTS ix_email_outbox_locked_until
    ON email_outbox(locked_until)
    WHERE locked_until IS NOT NULL;

CREATE TABLE IF NOT EXISTS crm_outbox (
    crm_id         uuid PRIMARY KEY,
    application_id uuid NOT NULL,
    payload        jsonb NOT NULL,
    status         text NOT NULL,
    attempt        int  NOT NULL DEFAULT 0,
    next_retry_at  timestamptz NULL,
    last_error     text NULL,
    created_at     timestamptz NOT NULL DEFAULT now(),
    updated_at     timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS ix_crm_outbox_pick
    ON crm_outbox(status, next_retry_at, created_at);

CREATE INDEX IF NOT EXISTS ix_crm_outbox_application
    ON crm_outbox(application_id);

CREATE TABLE IF NOT EXISTS integration_audit (
    audit_id    uuid PRIMARY KEY,
    system      text NOT NULL,
    entity_id   uuid NOT NULL,
    request     jsonb NULL,
    response    jsonb NULL,
    http_status int  NULL,
    created_at  timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS ix_integration_audit_system_created
    ON integration_audit(system, created_at DESC);

CREATE INDEX IF NOT EXISTS ix_integration_audit_entity
    ON integration_audit(entity_id);

COMMIT;
-- +goose StatementEnd


-- +goose Down
-- +goose StatementBegin
BEGIN;

DROP INDEX IF EXISTS ix_integration_audit_entity;
DROP INDEX IF EXISTS ix_integration_audit_system_created;
DROP TABLE IF EXISTS integration_audit;

DROP INDEX IF EXISTS ix_crm_outbox_application;
DROP INDEX IF EXISTS ix_crm_outbox_pick;
DROP TABLE IF EXISTS crm_outbox;

DROP INDEX IF EXISTS ix_email_outbox_locked_until;
DROP INDEX IF EXISTS ix_email_outbox_application;
DROP INDEX IF EXISTS ix_email_outbox_pick;
DROP TABLE IF EXISTS email_outbox;

DROP INDEX IF EXISTS ix_message_templates_active;
DROP INDEX IF EXISTS ux_message_templates_code;
DROP TABLE IF EXISTS message_templates;

DROP INDEX IF EXISTS ix_application_notes_application;
DROP TABLE IF EXISTS application_notes;

DROP INDEX IF EXISTS gin_applications_raw_row;
DROP INDEX IF EXISTS ix_applications_status_applied_at;
DROP INDEX IF EXISTS ix_applications_import;
DROP INDEX IF EXISTS ix_applications_candidate;
DROP INDEX IF EXISTS ix_applications_status;
DROP INDEX IF EXISTS ix_applications_applied_at;
DROP INDEX IF EXISTS ux_applications_external_key;
DROP TABLE IF EXISTS applications;

DROP INDEX IF EXISTS ix_candidate_contacts_candidate_primary;
DROP INDEX IF EXISTS ix_candidate_contacts_candidate;
DROP INDEX IF EXISTS ux_candidate_contacts_type_norm;
DROP TABLE IF EXISTS candidate_contacts;

DROP INDEX IF EXISTS ix_candidates_updated_at;
DROP INDEX IF EXISTS ix_candidates_name;
DROP TABLE IF EXISTS candidates;

DROP INDEX IF EXISTS ix_imports_status;
DROP INDEX IF EXISTS ix_imports_created_at;
DROP TABLE IF EXISTS imports;

COMMIT;
-- +goose StatementEnd
