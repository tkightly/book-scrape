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
     - TODO: decide whether to fail early if `Retry-After` is longer than the time left on `ctx`
     - TODO: decide what happens when a later chunk gives up after earlier chunks were already posted (Discord is sent before `SyncBooks`)
7. wait until the next schedule
