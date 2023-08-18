---
title: "Lookup Endpoint"
weight: 0
---

Returns the subdomains of the given domain in a JSON array or a newline delimetered string (to make scripting easier).

<img class="hidden-gif" src="/lookup.gif" alt="Lookup Gif">

<p class="p-center"><strong>Took less than a second for this endpoint to return 739 subdomains for <em>tesla.com</em>.</strong></p>

URL: `/api/lookup/{domain}`

The `domain` parameter must be a valid domain (eg.: `tesla.com`).

The `days` parameter can be used to finetune the result.

Check the details and try it out in the [documentation](https://columbus.elmasy.com/swagger/#/domain/get_api_lookup__domain_).

Example:

```bash
$ curl "https://columbus.elmasy.com/api/lookup/tesla.com"
["www", "mail", "shop", ...]

$ curl -H "Accept: text/plain" "https://columbus.elmasy.com/api/lookup/tesla.com"
www
mail
shop
...
```