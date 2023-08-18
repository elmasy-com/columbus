---
title: "History Endpoint"
weight: 1
---

Returns the DNS history of the given domain.

Columbus finds and stores the DNS records (eg.: A and AAAA) in the database to be used as a domain history tool.
If a known record found, updates the `time` field to the current time in unix timestamp.
If a new record found, appends to the records history.

Note: this is an **EXPERIMENTAL FEATURE**!

<img class="hidden-gif" src="/history.gif" alt="History Gif">

<p class="p-center"><strong>This endpoint will store every DNS record of <em>"columbus.elmasy.com"</em> and returns it is seconds.</strong></p>

URL: `/api/history/{domain}`

The `domain` parameter must be a valid domain (eg.: `example.com`). 

The `days` parameter can be used to finetune the result.

Check the details and try it out in the [documentation](https://columbus.elmasy.com/swagger/#/domain/get_api_history__domain_).

Example:

```bash
$ curl "https://columbus.elmasy.com/api/history/columbus.elmasy.com"
[{"type":1,"value":"142.132.164.231","time":1689723813},{"type":28,"value":"2a01:4f8:1c1e:eddd::1","time":1689723813}]
```

