CREATE TABLE IF NOT EXISTS users (
	id uuid,
	login VARCHAR,
	password VARCHAR,
	admin BOOLEAN,
	primary key (id)
);

CREATE TABLE IF NOT EXISTS balance (
	balanceid uuid,
	profileid uuid,
	operation double precision,
	operationtime timestamp DEFAULT NOW(),
	primary key (balanceid)
);