# Gitlab-CLI

Early stages, work in progress.

It is currently able to list users and groups recursively.

## Getting started

Install with `go get github.com/tommyknows/gitlab-cli` or through [packa](github.com/tommyknows/packa).

Then, setup gitlab-cli:

1. Create an instance: `gitlab-cli instance create <URL> <TOKEN>`
   Currently, it would _probably_ be enough to create an access token that
   has _read_api_, _read_user_ and _read_repository_ access. But for future
   functionality (and laziness), give it access to the whole API.
1. Ensure that the instance has been created by running `gitlab-cli instance list`
1. A default context has been generated automatically. However, you can create
   your own (setting the root to some kind of group, for example):
   `gitlab-cli context create <name> <instance> <group>`
1. Use it.

## Examples

Clone a whole group recursively:

```shell
gitlab-cli proj clone <group>
```

## Abbreviations

Because I'm lazy (and you probably are too), there are some abbreviations
defined on commands:

## Goals

- ✅Group-clone: check out a Gitlab-group recursively on the local Filesystem
- ❌Browse Merge Requests, merge them?, comment?

## TODOs

- Improve documentation / command help
- Improve existing code
  - Test the `Clone` function
  - Remove duplication from the cmd/ files
  - Probably move some code out of cmd/ and test it
