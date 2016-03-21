#git-mount

`git-mount` let's you mount your repo as a filesystem based on  a revision.

##Usage

Change to a directory that is an existing git repo. Once inside you can call
`git-mount` directly

```
git-mount HEAD
```

Or if `git-mount` is on your path you can just call it like an extension

```
git mount 2fdcb3ae
```

If only one argument is passed in `git-mount` treats that argument as a
[treeish](https://schacon.github.io/gitbook/4_git_treeishes.html). Based on
your current location in the repo it will mount all files and folders from that
level and deeper. `git-mount` will only ever descend files, never ascend, so if
you are in folder `foo` and folder `foo` is the top level of th repo the whole repo
will be mounted. If you go into `foo/bar` and call `git-mount {treeish}` then
only `bar` and its descendants will be mounted.

You can also pass a path to `git-mount`

```
git mount HEAD~2 public/img
```

This will tell `git-mount` to only mount the specified path and its
descendants.
