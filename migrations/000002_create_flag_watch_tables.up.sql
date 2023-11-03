CREATE TABLE IF NOT EXISTS flag_watch (
    "id" serial PRIMARY KEY,
    "submission_id" bigint NOT NULL,
    "problem_id" bigint NOT NULL,
    "user_id" bigint NOT NULL,
    "soft_delete" integer default 0,
    "mode" integer NOT NULL,
    "reason" text,
    "created_at" timestamptz,
    "updated_at" timestamptz,
    FOREIGN KEY (submission_id) REFERENCES submissions(id),
    FOREIGN KEY (problem_id) REFERENCES problems(id),
    FOREIGN KEY (user_id) REFERENCES users(id)
);
