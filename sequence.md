# Sequence flow

1. read data from sqlite
2. call algolia.worldofbooks.com to retrieve the list of books, paginating while page < nbPages
3. unmarshal the json
4. if any book is not in the sqlite, add it to a slice to be sent to discord and add it to the sqlite
5. any book that is in sqlite but not the most recent list, remove it from the sqlite
6. send to discord
   - TODO: handle Discord rate limits (HTTP 429): use the `X-RateLimit-*` / `Retry-After` response headers to wait and retry instead of failing the run
     - DONE: `sendWebRequestWithRetry` retries on 429 and has a test
     - DONE: parse `Retry-After` as a float (`ParseFloat`) so fractional seconds work; convert to `time.Duration` without truncating (e.g. `0.01`)
     - TODO: decide what to do when `Retry-After` is missing or unparseable (fail vs. default wait)
     - SKIPPED: cap the number of retries (decided not to; the context timeout is the only bound)
     - SKIPPED: test the cap (no cap, so nothing to test)
     - DONE: use a short `Retry-After` in the tests (e.g. `0.01`) so they run in milliseconds, and adjust the timeout formula
     - DONE: check the test fails when the retry is broken (temporarily break the 429 branch)
     - TODO: assert that every retry sends the same request body
     - TODO: fail early if `Retry-After` is longer than the time left on `ctx` (seen in the wild: `Retry-After=1679`, i.e. ~28 min, against the 1 minute context in `main`). Today it sleeps until the deadline and returns a bare `context deadline exceeded`
       - TODO: use `ctx.Deadline()` to compare (it returns `ok == false` when there is no deadline)
       - TODO: return a clear error naming the rate limit and the wait Discord asked for, not `context deadline exceeded`
       - TODO: add a test table row for a `Retry-After` longer than the context: expect an error quickly and only one request
     - TODO: fix the partial-failure loop: Discord is sent before `SyncBooks`, so if a later chunk fails, earlier chunks are already posted but never recorded. The next run treats every book as new, reposts the same chunks and fails again, so it will keep failing
       - TODO: decide the approach: record each chunk in sqlite right after it is posted, or sync before sending (missed notification instead of duplicates), or something else
       - TODO: log the chunk that failed and how many were posted before it
     - TODO: consider logging `X-RateLimit-Remaining` / `X-RateLimit-Reset-After` (and `X-RateLimit-Bucket`) per chunk to see how close a run comes to the limit
     - TODO: consider a summary log line at the end of a run (books sent, chunks, requests, rate limits hit)
7. wait until the next schedule
