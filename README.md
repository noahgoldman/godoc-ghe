# godoc-ghe

This is a tiny wrapper on top of [godoc](https://godoc.org/golang.org/x/tools/cmd/godoc) that periodically pulls all Go repositories from a GitHub Enterprise instance and serves docs for them.  

You need a [GitHub personal access token](https://help.github.com/articles/creating-a-personal-access-token-for-the-command-line/0) with the "repo" permission, and the public URL for GitHub Enterprise.

```
go get github.com/noahgoldman/godoc-ghe
godoc-ghe -gh-token <token> -gh-url <github public URL>
```
