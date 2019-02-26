---
title: "Profiles"
linkTitle: "Profiles"
weight: 70
---

Skaffold profiles allow you to define build, test and deployment
configurations for different contexts. Different contexts are typically different
environments in your app's lifecycle, like Production or Development. 

You can create profiles in the `profiles` section of `skaffold.yaml`.

For a detailed discussion on Skaffold configuration, see
[Skaffold Concepts](/docs/concepts/#configuration) and
[skaffold.yaml References](/docs/references/yaml).

## Profiles (`profiles`)

Each profile has four parts:

* Name (`name`): The name of the profile
* Build configuration (`build`)
* Test configuration (`test`)
* Deploy configuration (`deploy`)

Once activated, the specified `build`, `test` and `deploy` configuration
in the profile will override the `build`, `test` and `deploy` sections declared
in `skaffold.yaml`. The `build`, `test` and `deploy` configuration in the `profiles`
section use the same syntax as the `build`, `test` and `deploy` sections of
`skaffold.yaml`; for more information, see [Builders](/docs/how-tos/builders),
[Testers](/docs/how-tos/testers), and [Deployers](/docs/how-tos/deployers).

### Activation

You can activate profiles with the `-p` (`--profile`) parameter in the
`skaffold dev` and `skaffold run` commands.

```bash
skaffold run -p [PROFILE]
```

### Example

The following example, showcases a `skaffold.yaml` with one profile, `gcb`,
for building with Google Cloud Build:

{{% readfile file="samples/profiles/profiles.yaml" %}}

With no profile activated, Skaffold will build the artifact
`gcr.io/k8s-skaffold/skaffold-example` using local Docker daemon and deploy it
with `kubectl`.

However, if you run Skaffold with the following command:

```bash
skaffold dev -p gcb
```

Skaffold will switch to Google Cloud Build for building artifacts. Note that
since the `gcb` profile does not specify a deploy configuration, Skaffold will
continue using `kubectl` for deployments.
