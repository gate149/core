-- +goose Up
-- +goose StatementBegin
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

CREATE FUNCTION updated_at_update() RETURNS TRIGGER
    LANGUAGE plpgsql AS
$$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$;

CREATE FUNCTION check_max_problems_on_contest() RETURNS TRIGGER
    LANGUAGE plpgsql AS
$$
DECLARE
    max_problems_on_contest_count integer := 50;
BEGIN
    IF (SELECT count(*)
        FROM contest_problem
        WHERE contest_id = NEW.contest_id) >= (
           max_problems_on_contest_count
           ) THEN
        RAISE EXCEPTION 'Exceeded max problems for this contest';
    END IF;
    RETURN NEW;
END;
$$;

CREATE TABLE IF NOT EXISTS users
(
    id         uuid PRIMARY KEY,
    username   varchar(70) NOT NULL,
    role       varchar(60) NOT NULL DEFAULT 'user',
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),

    CHECK (length(username) != 0),
    CHECK (role = 'user' OR role = 'admin')
);

CREATE TRIGGER on_users_update
    BEFORE UPDATE
    ON users
    FOR EACH ROW
EXECUTE PROCEDURE updated_at_update();

CREATE TABLE IF NOT EXISTS problems
(
    id                 uuid PRIMARY KEY        DEFAULT uuid_generate_v4(),
    title              varchar(64)    NOT NULL,
    time_limit         integer        NOT NULL DEFAULT 1000,
    memory_limit       integer        NOT NULL DEFAULT 64,

    legend             varchar(10240) NOT NULL DEFAULT '',
    input_format       varchar(10240) NOT NULL DEFAULT '',
    output_format      varchar(10240) NOT NULL DEFAULT '',
    notes              varchar(10240) NOT NULL DEFAULT '',
    scoring            varchar(10240) NOT NULL DEFAULT '',

    legend_html        varchar(10240) NOT NULL DEFAULT '',
    input_format_html  varchar(10240) NOT NULL DEFAULT '',
    output_format_html varchar(10240) NOT NULL DEFAULT '',
    notes_html         varchar(10240) NOT NULL DEFAULT '',
    scoring_html       varchar(10240) NOT NULL DEFAULT '',

    owner_id           UUID           REFERENCES users (id) ON DELETE SET NULL,

    meta               jsonb          NOT NULL DEFAULT '{}',
    samples            jsonb          NOT NULL DEFAULT '[]',

    created_at         timestamptz    NOT NULL DEFAULT now(),
    updated_at         timestamptz    NOT NULL DEFAULT now(),

    CHECK (length(title) != 0),
    CHECK (memory_limit BETWEEN 4 and 1024),
    CHECK (time_limit BETWEEN 250 and 5000),
    CHECK ( (meta ->> 'count')::integer = jsonb_array_length(meta -> 'names') )
);

CREATE TRIGGER on_problems_update
    BEFORE UPDATE
    ON problems
    FOR EACH ROW
EXECUTE PROCEDURE updated_at_update();

CREATE TABLE IF NOT EXISTS contests
(
    id         uuid primary key     DEFAULT uuid_generate_v4(),
    title      varchar(64) NOT NULL,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),

    CHECK (length(title) != 0)
);

CREATE TRIGGER on_contests_update
    BEFORE UPDATE
    ON contests
    FOR EACH ROW
EXECUTE PROCEDURE updated_at_update();

CREATE TABLE IF NOT EXISTS contest_problem
(
    problem_id uuid    REFERENCES problems (id) ON DELETE SET NULL,
    contest_id uuid    REFERENCES contests (id) ON DELETE SET NULL,
    position   integer NOT NULL,
    UNIQUE (problem_id, contest_id),
    UNIQUE (contest_id, position),
    CHECK (position >= 0)
);

CREATE TRIGGER max_problems_on_contest_check
    BEFORE INSERT
    ON contest_problem
    FOR EACH STATEMENT
EXECUTE FUNCTION check_max_problems_on_contest();

CREATE TABLE IF NOT EXISTS contest_user
(
    user_id    uuid NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    contest_id uuid NOT NULL REFERENCES contests (id) ON DELETE CASCADE,
    UNIQUE (user_id, contest_id)
);

CREATE TABLE IF NOT EXISTS solutions
(
    id          uuid primary key          DEFAULT uuid_generate_v4(),
    contest_id  uuid             REFERENCES contests (id) ON DELETE SET NULL,
    problem_id  uuid             REFERENCES problems (id) ON DELETE SET NULL,
    user_id     uuid             REFERENCES users (id) ON DELETE SET NULL,
    solution    varchar(1048576) NOT NULL,
    state       integer          NOT NULL DEFAULT 1,
    score       integer          NOT NULL DEFAULT 0,
    penalty     integer          NOT NULL,
    time_stat   integer          NOT NULL DEFAULT 0,
    memory_stat integer          NOT NULL DEFAULT 0,
    language    integer          NOT NULL,
    updated_at  timestamptz      NOT NULL DEFAULT now(),
    created_at  timestamptz      NOT NULL DEFAULT now()
);

CREATE TRIGGER on_solutions_update
    BEFORE UPDATE
    ON solutions
    FOR EACH ROW
EXECUTE PROCEDURE updated_at_update();
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TRIGGER IF EXISTS on_solutions_update ON solutions;
DROP TABLE IF EXISTS solutions;
DROP TABLE IF EXISTS contest_user;
DROP TRIGGER IF EXISTS max_tasks_on_contest_check ON contest_problem;
DROP TABLE IF EXISTS contest_problem;
DROP TRIGGER IF EXISTS on_problems_update ON problems;
DROP INDEX IF EXISTS problem_title_trgm_idx;
DROP TABLE IF EXISTS problems;
DROP TRIGGER IF EXISTS on_contests_update ON contests;
DROP TABLE IF EXISTS contests;
DROP TRIGGER IF EXISTS on_users_update ON users;
DROP INDEX IF EXISTS users_username_trgm_idx;
DROP TABLE IF EXISTS users;
DROP FUNCTION IF EXISTS updated_at_update();
DROP FUNCTION IF EXISTS check_max_problems_on_contest();
-- +goose StatementEnd