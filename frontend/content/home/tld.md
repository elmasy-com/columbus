---
title: "TLD Endpoint"
weight: 3
---

Returns a list of all known Top Level Domains for the given domain.

<img class="hidden-gif" src="/tld.gif" alt="TLD Gif">

<p class="p-center"><strong>This endpoint can returns 377 different Top Level Domain for <em>google</em> in less than a second.</strong></p>

URL: `/api/tld/{domain}`

The domain parameter must be a valid Second Level Domain (eg.: example).

Learn more about this endpoint in the [documentation](https://columbus.elmasy.com/swagger/#/domain/get_api_tld__domain_).

Example:

```bash
$ curl "https://columbus.elmasy.com/api/tld/google"
["com", "hu", "co", ...]

$ curl -H "Accept: text/plain" "https://columbus.elmasy.com/api/tld/google"
com
hu
co
...
```