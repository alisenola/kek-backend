-- account
CREATE TABLE accounts (
	id serial PRIMARY KEY,
	username VARCHAR ( 50 ) UNIQUE NOT NULL,
	password VARCHAR ( 255 ) NOT NULL,
	email VARCHAR ( 255 ) UNIQUE NOT NULL,
	bio TEXT NULL,
	image VARCHAR ( 255 ) NULL,
	token VARCHAR ( 255 ) NULL,
	created_at TIMESTAMP NOT NULL,
	updated_at TIMESTAMP NOT NULL,
    last_login TIMESTAMP 
	disabled BOOLEAN NULL,
);

-- alert
CREATE TABLE alerts (
	id serial PRIMARY KEY,
	slug VARCHAR ( 100 ) NOT NULL,
	title VARCHAR ( 100 ) NOT NULL,
	body TEXT NULL,
	pair_address VARCHAR ( 100 ) NOT NULL,
	alert_type VARCHAR ( 100 ) NOT NULL,
	alert_value VARCHAR ( 20 ) NOT NULL,
	alert_option VARCHAR ( 20 ) NOT NULL,
	expiration_time TIMESTAMP,
	alert_actions VARCHAR ( 20 ) NOT NULL,
	alert_status VARCHAR ( 10 ) NOT NULL,
	account_id INTEGER NOT NULL,
	created_at TIMESTAMP NOT NULL,
	updated_at TIMESTAMP NOT NULL,
	deleted_at_unix BIGINT
);
