CREATE TABLE IF NOT EXISTS solutions (
    "id" serial PRIMARY KEY,
    "problem_id" bigint NOT NULL,
    "user_id" bigint NOT NULL,
    "code" text NOT NULL,
    "broadcast" integer DEFAULT 0,
    "created_at" timestamptz,
    "updated_at" timestamptz,
    FOREIGN KEY (problem_id) REFERENCES problems(id),
    FOREIGN KEY (user_id) REFERENCES users(id),
    UNIQUE(problem_id)
);
