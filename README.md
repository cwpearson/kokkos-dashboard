# [Dashboard for Kokkos](https://cwpearson.github.io/kokkos-dashboard/)

> [!WARNING]  
> Not an official Kokkos Ecosystem project.

Dashboard of Kokkos ecosystem activity.
Shows activity for the last two workdays for several repositories.


```
export KOKKOS_DASHBOARD_TOKEN=...

go mod tidy
go run *.go --fetch --render --serve
```


## Roadmap

- [x] Combine PRs and Issues, sort by most recent
- [x] Bit of JS to delete cards
  - [x] replace with `<details>`
- [x] more github.com links
- [x] Hide comments section if there are no comments
- [x] `owner-repo.html` -> `owner/repo/index.html`
- [x] Don't render after fetch errors
- [x] List commits
- [x] Favicon
  - [ ] make background transparent
- [x] tag draft PRs
- [ ] Interleave commits, comments, and events