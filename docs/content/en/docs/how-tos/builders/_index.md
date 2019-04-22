---
title: "Builders"
linkTitle: "Builders"
weight: 10
---

This page discusses how to set up Skaffold to use the tool of your choice
to build Docker images.

Skaffold supports the following tools to build your image:

* [Dockerfile](https://docs.docker.com/engine/reference/builder/) locally with Docker
* Dockerfile remotely with [Google Cloud Build](https://cloud.google.com/cloud-build/docs/)
* Dockerfile in-cluster with [Kaniko](https://github.com/GoogleContainerTools/kaniko)
* [Bazel](https://bazel.build/) locally
* [Jib](https://github.com/GoogleContainerTools/jib) Maven and Gradle projects locally
* [Jib](https://github.com/GoogleContainerTools/jib) remotely with [Google Cloud Build](https://cloud.google.com/cloud-build/docs/)

The `build` section in the Skaffold configuration file, `skaffold.yaml`,
controls how artifacts are built. To use a specific tool for building
artifacts, add the value representing the tool and options for using that tool
to the `build` section.

For a detailed discussion on Skaffold configuration, see
[Skaffold Concepts](/docs/concepts/#configuration) and
[skaffold.yaml References](/docs/references/yaml).

## Dockerfile locally with Docker

If you have [Docker Desktop](https://www.docker.com/products/docker-desktop)
installed, Skaffold can be configured to build artifacts with the local
Docker daemon.

By default, Skaffold connects to the local Docker daemon using
[Docker Engine APIs](https://docs.docker.com/develop/sdk/). Skaffold can, however,
be asked to use the [command-line interface](https://docs.docker.com/engine/reference/commandline/cli/)
instead. Additionally, Skaffold offers the option to build artifacts with
[BuildKit](https://github.com/moby/buildkit).

After the artifacts are successfully built, Docker images will be pushed
to the remote registry. You can choose to skip this step.

### Configuration

To use the local Docker daemon, add build type `local` to the `build` section
of `skaffold.yaml`. The following options can optionally be configured:

{{< schema root="LocalBuild" >}}

### Example

The following `build` section instructs Skaffold to build a
Docker image `gcr.io/k8s-skaffold/example` with the local Docker daemon:

{{% readfile file="samples/builders/local.yaml" %}}

Which is equivalent to:

{{% readfile file="samples/builders/local-full.yaml" %}}

## Dockerfile remotely with Google Cloud Build

[Google Cloud Build](https://cloud.google.com/cloud-build/) is a
[Google Cloud Platform](https://cloud.google.com) service that executes
your builds using Google infrastructure. To get started with Google
Build, see [Cloud Build Quickstart](https://cloud.google.com/cloud-build/docs/quickstart-docker).

Skaffold can automatically connect to Cloud Build, and run your builds
with it. After Cloud Build finishes building your artifacts, they will
be saved to the specified remote registry, such as
[Google Container Registry](https://cloud.google.com/container-registry/).

### Configuration

To use Cloud Build, add build type `googleCloudBuild` to the `build`
section of `skaffold.yaml`. The following options can optionally be configured:

{{< schema root="GoogleCloudBuild" >}}

### Example

The following `build` section, instructs Skaffold to build a
Docker image `gcr.io/k8s-skaffold/example` with Google Cloud Build:

{{% readfile file="samples/builders/gcb.yaml" %}}

## Dockerfile in-cluster with Kaniko

[Kaniko](https://github.com/GoogleContainerTools/kaniko) is a Google-developed
open-source tool for building images from a Dockerfile inside a container or
Kubernetes cluster. Kaniko enables building container images in environments
that cannot easily or securely run a Docker daemon.

Skaffold can help build artifacts in a Kubernetes cluster using the Kaniko
image; after the artifacts are built, kaniko can push them to remote registries.

### Configuration

To use Kaniko, add build type `kaniko` to the `build` section of
`skaffold.yaml`. The following options can optionally be configured:

{{< schema root="KanikoBuild" >}}

The `buildContext` can be either:

{{< schema root="KanikoBuildContext" >}}

### Example

The following `build` section, instructs Skaffold to build a
Docker image `gcr.io/k8s-skaffold/example` with Kaniko:

{{% readfile file="samples/builders/kaniko.yaml" %}}

## Jib Maven and Gradle locally

[Jib](https://github.com/GoogleContainerTools/jib#jib) is a set of plugins for
[Maven](https://github.com/GoogleContainerTools/jib/blob/master/jib-maven-plugin) and
[Gradle](https://github.com/GoogleContainerTools/jib/blob/master/jib-gradle-plugin)
for building optimized Docker and OCI images for Java applications
without a Docker daemon.

Skaffold can help build artifacts using Jib; Jib builds the container images and then
pushes them to the local Docker daemon or to remote registries as instructed by Skaffold.

### Configuration

To use Jib, add a `jibMaven` or `jibGradle` field to each artifact you specify in the
`artifacts` part of the `build` section. `context` should be a path to
your Maven or Gradle project.  

{{< alert title="Note" >}}
Your project must be configured to use Jib already.
{{< /alert >}}

The `jibMaven` type offers the following options:

{{< schema root="JibMavenArtifact" >}}

The `jibGradle` type offers the following options:

{{< schema root="JibGradleArtifact" >}}

### Example

See the [Skaffold-Jib demo project](https://github.com/GoogleContainerTools/skaffold/blob/master/examples/jib/)
for an example.

### Multi-Module Projects

Skaffold can be configured for _multi-module projects_ too. A multi-module project
has several _modules_ (Maven terminology) or _sub-projects_ (Gradle terminology) that
each produce a separate container image.

#### Maven

To build a Maven multi-module project, first identify the modules that should
produce a container image. Then for each such module:

  1. Create a Skaffold `artifact` in the `skaffold.yaml`:
     - Set the `artifact`'s `context` field to the root project location.
     - Add a `jibMaven` element and set its `module` field to the module's
       `:artifactId`, `groupId:artifactId`, or the relative path to the module
       _within the project_.
  2. Configure the module's `pom.xml` to bind either `jib:build` or `jib:dockerBuild` to
     the `package` phase as appropriate (see below).

This second step is necessary at the moment as Maven applies plugin goals specified
on the command-line, like `jib:build` or, to all modules and not just the modules
producing container images.
The situation is further complicated as Skaffold speeds up deploys to a local cluster,
such as `minikube`, by building and loading container images directly to the
local cluster's docker daemon (via `jib:dockerBuild` instead of `jib:build`),
thus saving a push and a pull of the image.
We plan to improve this situation [(#1876)](https://github.com/GoogleContainerTools/skaffold/issues/1876).

#### Gradle

To build a multi-module project with Gradle, specify each sub-project as a separate
Skaffold artifact. For each artifact, add a `jibGradle` field with a `project` field
containing the sub-project's name (the directory, by default). Each artifact's `context` field
should point to the root project location.

## Jib Maven and Gradle remotely with Google Cloud Build

{{% todo 1299 %}}

## Bazel locally

[Bazel](https://bazel.build/) is a fast, scalable, multi-language, and
extensible build system.

Skaffold can help build artifacts using Bazel; after Bazel finishes building
container images, they will be loaded into the local Docker daemon.

### Configuration

To use Bazel, `bazel` field to each artifact you specify in the
`artifacts` part of the `build` section, and use the build type `local`.
`context` should be a path containing the bazel files
(`WORKSPACE` and `BUILD`). The following options can optionally be configured:

{{< schema root="BazelArtifact" >}}

### Example

The following `build` section instructs Skaffold to build a
Docker image `gcr.io/k8s-skaffold/example` with Bazel:

{{% readfile file="samples/builders/bazel.yaml" %}}
