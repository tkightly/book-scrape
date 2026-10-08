# Sequence flow

1. read data from sqlite
2. call algolia.worldofbooks.com to retrieve the list of books, paginating while page < nbPages
3. unmarshal the json
4. if any book is not in the sqlite, add it to a slice to be sent to discord and add it to the sqlite
5. any book that is in sqlite but not the most recent list, remove it from the sqlite
6. send to discord
   - TODO: handle Discord rate limits (HTTP 429): use the `X-RateLimit-*` / `Retry-After` response headers to wait and retry instead of failing the run
     - TODO: fix the partial-failure loop: Discord is sent before `SyncBooks`, so if a later chunk fails, earlier chunks are already posted but never recorded. The next run treats every book as new, reposts the same chunks and fails again, so it will keep failing
       - TODO: decide the approach: record each chunk in sqlite right after it is posted, or sync before sending (missed notification instead of duplicates), or something else
       - TODO: log the chunk that failed and how many were posted before it
     - TODO: consider logging `X-RateLimit-Remaining` / `X-RateLimit-Reset-After` (and `X-RateLimit-Bucket`) per chunk to see how close a run comes to the limit
     - TODO: consider a summary log line at the end of a run (books sent, chunks, requests, rate limits hit)
7. wait until the next schedule
