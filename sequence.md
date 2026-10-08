# Sequence flow

1. read data from sqlite
2. call algolia.worldofbooks.com to retrieve the list of books, paginating while page < nbPages
   - TODO: handle Algolia rate limits (HTTP 429): `algolia.Search` pages serially with no pause and fails the whole run on any non-2xx. Seen after several runs in a few minutes (page 13 for one author, then page 0 for another). Read and log the response status, body and headers; consider a short pause between pages, a retry with backoff like the Discord one, and a larger `hitsPerPage` to cut the number of requests (check the Algolia docs for the limit)
3. unmarshal the json
4. if any book is not in the sqlite, add it to a slice to be sent to discord and add it to the sqlite
5. any book that is in sqlite but not the most recent list, remove it from the sqlite
6. send to discord
   - TODO: find out which Discord limit we are hitting: log the 429 response body (`message`, `retry_after`, `global`) and headers (`X-RateLimit-Scope`, `X-RateLimit-Remaining`, `X-RateLimit-Reset-After`, `X-RateLimit-Bucket`) in the "rate limited" log line. Observed twice: 10 chunks post, then a ~28 minute `Retry-After`, even 11 minutes apart
   - TODO: decide whether a Discord rate limit is an error (`log.Fatal`, exit 1) or an expected outcome, since the remaining chunks are sent by later runs
   - TODO: consider a summary log line at the end of a run (books sent, chunks, requests, rate limits hit)
7. wait until the next schedule
