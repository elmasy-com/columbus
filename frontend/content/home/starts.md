---
title: "Starts Endpoint"
weight: 2
---

Returns a list of all known Second Level Domains thats start with the given domain in a JSON array or a newline delimetered string.

<img class="hidden-gif" src="/starts.gif" alt="Starts Gif">

<p class="p-center"><strong>This endpoint can returns 1809 domains that starts with <em>"reddit"</em> in less than 4 seconds.</strong></p>

URL: `/api/starts/{domain}`

The `domain` parameter must be at least five character long, valid Second Level Domain (eg.: `example`). 

Read the details and try it yourself in the [documentation](https://columbus.elmasy.com/swagger/#/domain/get_api_starts__domain_).

Example:

```bash
$ curl "https://columbus.elmasy.com/api/starts/reddit"
["reddit", "redditmedia", "redditstatistic", ...]

$ curl -H "Accept: text/plain" "https://columbus.elmasy.com/api/starts/reddit"
reddit
redditmedia
redditstatistic
...
```