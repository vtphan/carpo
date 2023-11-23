CREATE TABLE IF NOT EXISTS student_checkout_status (
    "id" serial PRIMARY KEY,
    "user_id" bigint NOT NULL,
    "problem_id" bigint NOT NULL,
    "created_at" timestamptz,
    "updated_at" timestamptz,
    FOREIGN KEY (user_id) REFERENCES users(id),
    FOREIGN KEY (problem_id) REFERENCES problems(id)
);
