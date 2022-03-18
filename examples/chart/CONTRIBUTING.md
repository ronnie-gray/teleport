# Contributing to Teleport Helm charts

Firstly, thanks for considering a contribution to Teleport's Helm charts.

## A couple of brief warnings

Please note that we won't accept contributions that are particularly esoteric, difficult to use or poorly implemented.
Our goal is to:

- keep the charts easy to use
- keep all functionality relevant to a broad audience
- always use sane defaults which are right for most deployments
- require as few values changes as possible for everyday usage

If your functionality is only really useful to you, it's best to keep it on your own fork and deploy from there.

Sometimes Teleport staff may take over your PR and make changes, or implement it in a slightly different way. We will
make sure that you still get credited in the final commit if this happens.

## Guidelines

Here is a list of things that you should do to make sure to do in order to get a smooth PR review process with minimal
changes required:

1) Add a linter file which includes examples for any new values you add under the `.lint/` directory for the
appropriate chart. The linter will check this during CI and make sure the values are correctly formatted, along
with your chart changes. The file should contain all necessary values to deploy a reference install.

2) Add unit tests for your functionality under the `tests/` directory for the appropriate chart, particularly if you're
adding new values. Make sure that all functionality is tested, so we can be sure that it works as intended for every use
case. A good tip is to use your newly added linter file to set values appropriate for your test.

3) Add any new values at the correct location in the `values.schema.json` file for the appropriate chart. This
will ensure that Helm is able to validate values at install-time and can prevent users from making easy mistakes.

4) Document any new values or changes to existing behaviour in the [chart reference](../../docs/pages/kubernetes-access/helm/reference.mdx).

5) Run `make lint-helm test-helm` from the root of the repo before raising your PR.
You will need `yamllint`, `helm` and [helm3-unittest](https://github.com/vbehar/helm3-unittest) installed locally.

`make -C build.assets lint-helm test-helm` will run these via Docker if you'd prefer not to install locally.

6) If you get a snapshot error during your testing, you should verify that your changes intended to alter the output,
then run `make update-helm-snapshots` to update the snapshots and commit these changes along with your PR.

Again, `make -C build.assets update-helm-snapshots` will run this via Docker.

7) Document the changes you've made in the PR comment and add @webvictim as a reviewer.

Thanks!