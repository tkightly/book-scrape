# Sequence flow

1. read data from sqlite
2. call algolia.worldofbooks.com to retrieve the list of books, paginating while page < nbPages
3. unmarshal the json
4. if any book is not in the sqlite, add it to a slice to be sent to discord and add it to the sqlite
5. any book that is in sqlite but not the most recent list, remove it from the sqlite
6. send to discord
   - TODO: handle Discord rate limits (HTTP 429): use the `X-RateLimit-*` / `Retry-After` response headers to wait and retry instead of failing the run
     - TODO: fix the partial-failure loop: Discord is sent before `SyncBooks`, so if a later chunk fails, earlier chunks are already posted but never recorded. The next run treats every book as new, reposts the same chunks and fails again, so it will keep failing
       - DECIDED: `run()` does the chunking: for each chunk, post it, then `SyncBooks` for that chunk, and stop at the first failure. At most one chunk (10 books) can be repeated after a failure or crash. Rejected: sync before sending (loses notifications), one `SyncBooks` at the end (a crash or an expired `ctx` loses all progress), callback into `SendWebhook`, re-running `ListBooks` + `doDiff` per chunk
       - TODO: export the chunk size as a constant from `discord` (embeds per message) and use it in `run()`; keep chunking inside `SendWebhook` as a safety net
       - TODO: pass the header `content` string into `SendWebhook`; `run()` builds it once from the total count and gives it only to the first chunk (empty string for the rest). Decide where the author-name formatting (`buildAuthorNames`) lives
       - TODO: pass `removed` to the first `SyncBooks` call only, or to every call (safe either way: the delete of a missing row does nothing)
       - TODO: update `TestSendWebhook` for the new responsibilities (header and chunking moved to `run()`)
       - TODO: log the chunk that failed and how many were posted before it
     - TODO: consider logging `X-RateLimit-Remaining` / `X-RateLimit-Reset-After` (and `X-RateLimit-Bucket`) per chunk to see how close a run comes to the limit
     - TODO: consider a summary log line at the end of a run (books sent, chunks, requests, rate limits hit)
7. wait until the next schedule
