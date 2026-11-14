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

## Container image & Helm chart

A `vX.Y.Z` tag publishes, besides the release binaries:

- the image `ghcr.io/mycroft/rss-to-bluesky:X.Y.Z` (amd64, arm64; distroless,
  runs as UID 65532). Its working directory is `/app`, so the database is
  `/app/database/db`: mount a volume on `/app/database`;
- the chart `oci://ghcr.io/mycroft/charts/rss-to-bluesky`, version and
  appVersion `X.Y.Z`.

The chart runs the tool as a CronJob (every 5 minutes by default, never two
runs at once) with the database on a PersistentVolumeClaim, kept on uninstall.
It reads the credentials from an existing Secret, keys `bluesky-user` and
`bluesky-password`:

```sh
$ helm install rss-to-bluesky oci://ghcr.io/mycroft/charts/rss-to-bluesky \
    --set existingSecret=rss-to-bluesky
```

See `charts/rss-to-bluesky/values.yaml` for the rest (schedule, extra flags,
storage class, resources). Lint it with:

```sh
$ for v in charts/rss-to-bluesky/ci/*-values.yaml; do helm lint --strict charts/rss-to-bluesky -f $v; done
```

### Moving an existing database

The database records every item already posted: a run against an empty one
posts the whole feed again. To move an install to the chart, keep its
database:

1. Install the chart with `--set suspend=true`, so nothing runs yet.
2. Stop the old install (its cron entry), so the database stops changing.
3. Copy the database into the claim. The image has no shell, so use a helper
   pod:

   ```sh
   $ kubectl -n <ns> run db-copy --image=busybox --restart=Never \
       --overrides='{"spec":{"containers":[{"name":"db-copy","image":"busybox","command":["sleep","3600"],"volumeMounts":[{"name":"db","mountPath":"/db"}]}],"volumes":[{"name":"db","persistentVolumeClaim":{"claimName":"rss-to-bluesky"}}]}}'
   $ kubectl -n <ns> cp database/db db-copy:/db/db
   $ kubectl -n <ns> exec db-copy -- chown 65532:65532 /db/db
   $ kubectl -n <ns> exec db-copy -- chmod 600 /db/db
   $ kubectl -n <ns> delete pod db-copy
   ```

4. Set `suspend: false`, and check the first run's logs: it should only
   post the items that appeared since the old install stopped.
