-- initial schema

CREATE TABLE bookmark (
    id          INTEGER NOT NULL PRIMARY KEY,
    name        TEXT,
    url         TEXT,
    date        DATETIME,
    description  TEXT
);

CREATE TABLE tag (
    id      INTEGER NOT NULL PRIMARY KEY,
    name    TEXT UNIQUE
);

CREATE TABLE tag_bookmark (
    tag         INTEGER NOT NULL,
    bookmark    INTEGER NOT NULL,
    PRIMARY KEY (tag, bookmark),
    FOREIGN KEY (tag) REFERENCES tag(id) ON DELETE CASCADE,
    FOREIGN KEY (bookmark) REFERENCES bookmark(id) ON DELETE CASCADE
);