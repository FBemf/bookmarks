# A simple service for storing your bookmarks

I wrote this app because I didn't like the maximalism of Pocket or similar services.
I just wanted somewhere to aggregate every website I thought I'd ever want to look at again, with no bells & whistles.
I also made it to try out [hotwire][hotwire], which I found very pleasant to use.

[hotwire]: https://hotwire.dev/

![screenshot](screenshot.png)

As a web app, it's mostly self-explanatory.
Serve it with the `serve` command.
Adding users and changing their passwords is done with the `user` command.
Best practice, though, is to do that with `./set_password.sh <USER>`, which interactively prompts for the password so that it stays out of the shell history.

## Features

- Tag your bookmarks
- Search, filter by tags, or do both at the same time
- Includes a javascript bookmarklet for easy bookmarking (found on the API keys page)
- Compiles to just one binary, including sqlite driver

## API

The only thing that isn't clear from the UI is the API, which has two endpoints:

- `POST /api/bookmark` takes a json of the format
`{"name": "Site Name", "url": "https://example.com", "description": "A description", "tags": ["tag1", "tag2"]}`
and adds that website as a bookmark.
- `GET /api/export` returns a json document full of all the bookmarks in the database.
There is currently no way to import from such a document; at the moment the only way to import bookmarks is to write them into the sqlite database using a script.
This is mostly just for backups.