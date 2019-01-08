CREATE EXTENSION IF NOT EXISTS CITEXT;

DROP TABLE IF EXISTS public.users CASCADE;
DROP TABLE IF EXISTS public.forums CASCADE;
DROP TABLE IF EXISTS public.threads CASCADE;
DROP TABLE IF EXISTS public.posts CASCADE;
DROP TABLE IF EXISTS public.votes CASCADE;
DROP TABLE IF EXISTS public.forum_users CASCADE;


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
	author     CITEXT REFERENCES users(nickname),
	forum    CITEXT REFERENCES forums(slug),
	votes    BIGINT DEFAULT 0
);

CREATE UNIQUE INDEX threads_id_forum_idx on threads(id, forum);            -- ?
CREATE UNIQUE INDEX threads_slug_id_forum_idx on threads(slug, id, forum); -- ? 

-- GetThreads - ?
CREATE UNIQUE INDEX threads_forum_created_idx on threads(forum, created); -- ? 


CREATE TABLE posts (
  id        SERIAL	   NOT NULL PRIMARY KEY,
  author    CITEXT	   NOT NULL REFERENCES users(nickname),
  created   TIMESTAMP WITH TIME ZONE,
  forum     CITEXT     REFERENCES forums(slug),
  isEdited  BOOLEAN	   DEFAULT FALSE,
  message   TEXT	   NOT NULL,
  parent    INTEGER	   DEFAULT 0,
  thread    INTEGER	   NOT NULL REFERENCES threads(id),
  
  path      INTEGER	ARRAY
);

CREATE UNIQUE INDEX posts_thread_id_idx on posts(thread, id, path);        -- ? 

CREATE UNIQUE INDEX posts_thread_path_idx on posts(thread, path);        -- ? 

CREATE UNIQUE INDEX posts_id_root_idx on posts(id, (path[1])); -- ??

CREATE UNIQUE INDEX posts_root_idx on posts(thread, (path[1]) desc, path); -- ??

CREATE UNIQUE INDEX posts_parent_thread_root_id_idx on posts(parent, thread, (path[1]), id); -- ??

CREATE UNIQUE INDEX posts_parent_thread_id_idx on posts(parent, thread, id); -- ??


CREATE TABLE votes (
  id        SERIAL      NOT NULL PRIMARY KEY,
  nickname  CITEXT      NOT NULL REFERENCES users(nickname),
  voice     INTEGER,
  thread    INTEGER     NOT NULL REFERENCES threads(id),
  
	UNIQUE(nickname, thread)
);


CREATE TABLE forum_users
(
  username  CITEXT	 NOT NULL   REFERENCES users(nickname),
  forum     CITEXT   NOT NULL   REFERENCES forums(slug) ,

  UNIQUE(forum, username)
);

CREATE INDEX IF NOT EXISTS forum_users_username_idx ON forum_users(username);
CREATE INDEX IF NOT EXISTS forum_users_forum_idx ON forum_users(forum);




-- --------------------------- Triggers ---------------------------

CREATE FUNCTION fix_path() RETURNS trigger AS $fix_path$
BEGIN
  new.path := array_append((SELECT path from posts WHERE thread = new.thread AND id = new.parent), new.id);
  insert into forum_users (forum, username) values (new.forum, new.author) ON conflict (forum, username) do nothing;
  RETURN new;
END;
$fix_path$ LANGUAGE plpgsql;


CREATE TRIGGER fix_path BEFORE INSERT OR UPDATE ON posts FOR EACH ROW EXECUTE PROCEDURE fix_path();
