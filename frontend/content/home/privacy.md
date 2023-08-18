---
title: "Privacy Policy"
weight: 5
---

- Every request to the server logged and stored for 30 days.
- The website is using the [privacy-friendly](https://plausible.io/privacy-focused-web-analytics) Plausible to collect analytics.
- Successful queries to `/api/lookup` saves **only** the `domain` parameter for further process. See the relevant code [here](https://github.com/elmasy-com/columbus-server/blob/dae62d64fe85391e4257a219ea5324e782fb1cc2/server/lookup.go#L61).
- Failed queries to `/api/lookup` saves **only** the `domain` parameter for further process. See the relevant code [here](https://github.com/elmasy-com/columbus-server/blob/dae62d64fe85391e4257a219ea5324e782fb1cc2/server/lookup.go#L48).
