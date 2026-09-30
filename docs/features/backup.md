English | [简体中文](backup.zh-CN.md)

# System Backup and WebDAV

SublinkPro can upload its existing system backup directly to WebDAV and restore the current instance from a remote ZIP backup. The feature reuses the existing backup format and database migration workflow.

## Usage

1. Sign in as an administrator and open **User Center → System Backup**.
2. Enter the WebDAV URL, username, password or app password, and remote directory.
3. Select **Test connection** to confirm that the account can read the remote directory. Missing directories are created during upload when possible.
4. Select **Save settings**.
5. Optionally enable **scheduled backup** and choose a 5-field cron expression, for example `0 3 * * *` for 03:00 every day.
6. Select **Back up to WebDAV now** to create and upload a ZIP immediately.
7. Choose a ZIP from the remote backup list and confirm the overwrite warning to restore it.

Example WebDAV URL:

```text
https://dav.example.com/remote.php/dav/files/username
```

Example remote directory:

```text
SublinkPro/backups
```

Do not put credentials, query parameters, or fragments in the URL. Missing remote directories are created one level at a time with WebDAV `MKCOL` when the server permits it.

## Backup contents

The ZIP contains the same data as the local System Backup action in the avatar menu:

- `db/`: the current data directory;
- `template/`: the template directory;
- the re-downloadable `GeoLite2-City.mmdb` file is excluded.

WebDAV system backup currently supports SQLite only. With MySQL or PostgreSQL, **Back up to WebDAV now** and scheduled backups are rejected because the local data directory does not contain business data from the remote database server.

## Scheduled backup

When scheduled backup is enabled, SublinkPro registers a system cron job that creates the same ZIP used by the manual upload action and sends it to the configured WebDAV directory.

- The schedule uses a standard 5-field cron expression (`minute hour day month weekday`).
- Saving WebDAV settings hot-reloads the job without restarting the process.
- Restoring a backup preserves the current instance's WebDAV connection and schedule settings.
- Overlapping runs are skipped if a previous scheduled backup is still uploading.
- Progress appears in the task center as a `webdav_backup` task.

## Restore behavior

A remote restore first downloads the selected ZIP to the server's temporary migration directory and then reuses the database migration task:

- current business data is overwritten;
- the template directory in the ZIP is restored;
- restoring access keys and subscription access logs is optional;
- the current instance's JWT, Cloudflare Tunnel, and WebDAV connection settings are preserved;
- an instance restart may be required afterward.

Back up the current instance before restoring. Progress and results are available in the task center.

## Security

- WebDAV settings and all backup endpoints are administrator-only and disabled in demo mode.
- The WebDAV password is encrypted with the instance `SUBLINK_API_ENCRYPTION_KEY`; API responses never return it in plaintext.
- Backups may contain account data, access keys, node credentials, private keys, and other sensitive information.
- HTTPS is required by default. HTTP is accepted only after an explicit opt-in and should be used only on a trusted isolated network.
- Loopback, private, and reserved addresses are rejected by default. Explicitly enable private-network access only for a trusted private NAS.
- The client never follows redirects; same-origin `MKCOL` collection redirects are accepted without a follow-up request. Remote filenames must be a single `.zip` filename.
- Remote uploads and downloads are limited to 1 GiB per backup. The UI lists at most the newest 200 ZIP files.

## Current scope

The current release supports manual connection testing, upload, listing, restore, and scheduled backups. Remote deletion and automatic retention policies are not included yet.
