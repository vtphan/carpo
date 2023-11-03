
CREATE TABLE IF NOT EXISTS users (
   "id" serial PRIMARY KEY,
   "name" VARCHAR (128) NOT NULL,
   "user_uuid" VARCHAR (36) NOT NULL,
   "role" integer NOT NULL
);

CREATE TABLE IF NOT EXISTS problems (
    "id" serial PRIMARY KEY,
    "user_id" bigint NOT NULL,
    "question" text NOT NULL,
    "format" VARCHAR(10) NOT NULL,
    "lifetime" timestamptz NOT NULL,
    "status" integer NOT NULL,
    "created_at" timestamptz,
    "updated_at" timestamptz,
    FOREIGN KEY (user_id) REFERENCES users(id)

);

CREATE TABLE IF NOT EXISTS submissions (
    "id" serial PRIMARY KEY,
    "problem_id" bigint NOT NULL,
    "user_id" bigint NOT NULL,
    "message" text NOT NULL,
    "code" text NOT NULL,
    "is_snapshot" integer NOT NULL,
    "status" integer NOT NULL,
    "created_at" timestamptz,
    "updated_at" timestamptz,
    FOREIGN KEY (problem_id) REFERENCES problems(id),
    FOREIGN KEY (user_id) REFERENCES users(id)

);

CREATE TABLE IF NOT EXISTS grades (
    "id" serial PRIMARY KEY,
    "user_id" bigint NOT NULL,
    "submission_id" bigint NOT NULL,
    "score" integer NOT NULL,
    "code" text NOT NULL,
    "comment" text NOT NULL,
    "status" integer NOT NULL,
    "has_feedback" integer NOT NULL,
    "feedback_at" timestamptz,
    "created_at" timestamptz,
    "updated_at" timestamptz,
    FOREIGN KEY (submission_id) REFERENCES submissions(id),
    FOREIGN KEY (user_id) REFERENCES users(id),
    CONSTRAINT UC_Grade UNIQUE (user_id, submission_id)
);

CREATE TABLE IF NOT EXISTS student_problem_status (
    "id" serial PRIMARY KEY,
    "user_id" bigint NOT NULL,
    "problem_id" bigint NOT NULL,
    "problem_status" integer NOT NULL,
    "created_at" timestamptz,
    "updated_at" timestamptz,
    FOREIGN KEY (user_id) REFERENCES users(id),
    FOREIGN KEY (problem_id) REFERENCES problems(id)
);
