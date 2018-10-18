DROP TABLE IF EXISTS public.users CASCADE;
DROP TABLE IF EXISTS public.forums CASCADE;
DROP TABLE IF EXISTS public.threads CASCADE;

CREATE TABLE users (
	about TEXT NOT NULL,
	email CITEXT NOT NULL UNIQUE,
	fullname TEXT NOT NULL,
	nickname CITEXT COLLATE ucs_basic NOT NULL UNIQUE,
	id BIGSERIAL PRIMARY KEY
);

CREATE TABLE forums (
	id BIGSERIAL PRIMARY KEY,
	slug CITEXT NOT NULL UNIQUE,
	title CITEXT,
	author CITEXT REFERENCES users(nickname),
	threads INTEGER DEFAULT 0,
	posts INTEGER DEFAULT 0
);

CREATE TABLE threads
(
	id         BIGSERIAL PRIMARY KEY,
	slug       CITEXT UNIQUE,
	created    TIMESTAMP WITH TIME ZONE,
	message    TEXT,
	title      TEXT,
	author     CITEXT REFERENCES users (nickname),
	forum    CITEXT REFERENCES forums(slug),
	votes    BIGINT DEFAULT 0
);