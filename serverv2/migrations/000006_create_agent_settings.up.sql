CREATE TABLE IF NOT EXISTS agent_settings (
    "id" serial PRIMARY KEY,
    "name" text NOT NULL,
    "description" text NOT NULL,
    "prompt" text NOT NULL,
    "model" text NOT NULL,
    "is_active" integer NOT NULL,
    "created_at" timestamptz,
    "updated_at" timestamptz
);