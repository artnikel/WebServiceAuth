CREATE TABLE IF NOT EXISTS users (
	id uuid,
	login VARCHAR,
	password VARCHAR,
	admin BOOLEAN,
	primary key (id)
);