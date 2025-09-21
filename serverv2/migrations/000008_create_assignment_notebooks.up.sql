CREATE TABLE IF NOT EXISTS assignment_notebooks (
    "id" serial PRIMARY KEY,
    "notebook_uuid" VARCHAR (36) NOT NULL,
    "title" VARCHAR (128) NOT NULL,
    "mode" integer NOT NULL,
    "path" text NOT NULL,
    "available_till" timestamptz,
    "end_time" timestamptz,
    "user_id" bigint NOT NULL,
    "created_at" timestamptz,
    "updated_at" timestamptz,
    FOREIGN KEY (user_id) REFERENCES users(id),
    CONSTRAINT UC_title_mode UNIQUE (title, mode)
);

CREATE TABLE IF NOT EXISTS students_notebooks_submissions (
    "id" serial PRIMARY KEY, 
    "notebook_id" bigint NOT NULL,
    "title" VARCHAR (128) NOT NULL,
    "path" text NOT NULL,
    "submission_status" bigint NOT NULL,
    "submitted_at" timestamptz,
    "file_created_at" timestamptz,
    "user_id" bigint NOT NULL,
    "created_at" timestamptz,
    "updated_at" timestamptz,
    FOREIGN KEY (notebook_id) REFERENCES assignment_notebooks(id),
    FOREIGN KEY (user_id) REFERENCES users(id)
);