DROP TABLE session;

CREATE TABLE session (
    id          INTEGER PRIMARY KEY,
    user        INTEGER NOT NULL,
    timestamp   TIMESTAMP NOT NULL,
    cookie      BLOB NOT NULL,
    csrf        BLOB NOT NULL,
    FOREIGN KEY (user) REFERENCES user(id) ON DELETE CASCADE
);

CREATE INDEX session__cookie ON session(cookie);