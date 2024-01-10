CREATE TABLE IF NOT EXISTS tags (
    "id" serial PRIMARY KEY,
    "name" text NOT NULL,
    "mode" bigint NOT NULL,
    "status" bigint NOT NULL,
    "user_id" bigint NOT NULL,
    "created_at" timestamptz,
    "updated_at" timestamptz,
    FOREIGN KEY (user_id) REFERENCES users(id),
    CONSTRAINT UC_Tag UNIQUE (name, mode, status)

);

CREATE TABLE IF NOT EXISTS problem_tag (
    "id" serial PRIMARY KEY,
    "problem_id" bigint NOT NULL,
    "tag_id" bigint NOT NULL,
    "notes" text,
    "user_id" bigint NOT NULL,
    "created_at" timestamptz,
    "updated_at" timestamptz,
    FOREIGN KEY (problem_id) REFERENCES problems(id),
    FOREIGN KEY (user_id) REFERENCES users(id),
    FOREIGN KEY (tag_id) REFERENCES tags(id)
);

CREATE TABLE IF NOT EXISTS submission_tag (
    "id" serial PRIMARY KEY,
    "submission_id" bigint NOT NULL,
    "tag_id" bigint NOT NULL,
    "notes" text,
    "user_id" bigint NOT NULL,
    "created_at" timestamptz,
    "updated_at" timestamptz,
    FOREIGN KEY (user_id) REFERENCES users(id),
    FOREIGN KEY (submission_id) REFERENCES submissions(id),
    FOREIGN KEY (tag_id) REFERENCES tags(id)
);

-- Unique constrain