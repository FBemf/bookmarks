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
    name    TEXT
);

CREATE TABLE tag_bookmark (
    tag         INTEGER,
    bookmark    INTEGER,
    FOREIGN KEY (tag) REFERENCES tag(id),
    FOREIGN KEY (bookmark) REFERENCES bookmark(id)
);

CREATE TABLE unread_bookmark (
    bookmark    INTEGER PRIMARY KEY,
    FOREIGN KEY (bookmark) REFERENCES bookmark(id)
);