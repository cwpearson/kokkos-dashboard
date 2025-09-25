# [kokkos-dashboard](https://cwpearson.github.io/kokkos-dashboard/)

> [!WARNING]  
> Not an official Kokkos Ecosystem project.

Dashboard of Kokkos ecosystem activity.


```
export KOKKOS_DASHBOARD_TOKEN=...

go mod tidy
go run *.go --fetch --render --serve
```


## Roadmap

- [x] Combine PRs and Issues, sort by most recent
- [x] Bit of JS to delete cards
- [ ] more github.com links
- [x] Hide comments section if there are no comments
- [x] `owner-repo.html` -> `owner/repo/index.html`