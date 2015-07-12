DROP TABLE IF EXISTS "alias";
DROP TABLE IF EXISTS "person";

CREATE TABLE "person" (
    id  SERIAL PRIMARY KEY,
    created TIMESTAMP NOT NULL DEFAULT NOW(),
    name    TEXT NOT NULL UNIQUE,
    phone   TEXT NOT NULL,
    last_summon TIMESTAMP
);

CREATE TABLE "alias" (
    id  SERIAL PRIMARY KEY,
    created TIMESTAMP NOT NULL DEFAULT NOW(),
    name    TEXT NOT NULL,
    person  INTEGER NOT NULL REFERENCES person(id)
);

