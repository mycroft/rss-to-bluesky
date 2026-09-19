# rss-to-bluesky

I'll be honest with you: I need to refactor a lot of things in this tool. But it works!

You can find it live on [Bluesky](https://bsky.app/profile/lobsters-feed.bsky.social).

## Build & run

```sh
$ go build ./cmd/rss-to-bluesky/
$ export BLUESKY_USER="..."
$ export BLUESKY_PASS="..."
$ ./rss-to-bluesky
```

The state database is created at `./database/db`, relative to the working
directory. It stores the Bluesky session tokens, so it is created readable by
its owner only. Earlier versions created it world-readable: if you already have
one, tighten it once with

```sh
$ chmod 600 database/db
```
