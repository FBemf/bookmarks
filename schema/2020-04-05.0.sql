-- initial schema

CREATE TABLE bookmark (
    id          INTEGER PRIMARY KEY,
    name        TEXT,
    url         TEXT,
    date        DATETIME,
    description  TEXT
);

CREATE TABLE tag (
    id      INTEGER PRIMARY KEY,
    name    TEXT UNIQUE
);

CREATE TABLE tag_bookmark (
    tag         INTEGER NOT NULL,
    bookmark    INTEGER NOT NULL,
    PRIMARY KEY (tag, bookmark),
    FOREIGN KEY (tag) REFERENCES tag(id) ON DELETE CASCADE,
    FOREIGN KEY (bookmark) REFERENCES bookmark(id) ON DELETE CASCADE
);

CREATE TABLE user (
    id            INTEGER PRIMARY KEY,
    username      TEXT UNIQUE,
    password_hash BLOB,
    salt          BLOB
);

CREATE TABLE session (
    id          INTEGER PRIMARY KEY,
    user        INTEGER NOT NULL,
    timestamp   TIMESTAMP NOT NULL,
    cookie      BLOB NOT NULL,
    FOREIGN KEY (user) REFERENCES user(id) ON DELETE CASCADE
);