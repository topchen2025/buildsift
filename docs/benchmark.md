# Initial validation set

Snapshot: 2026-07-20

This is a small, transparent validation set for BuildSift's current rule engine. It is **not** a claim of real-world accuracy, a held-out benchmark, or a comparison with another tool.

The 20 inputs are short, sanitized reproductions of common failure shapes. They were minimized or rewritten from the linked public reports so that this repository does not redistribute full build logs, personal paths, credentials, or unrelated source code. A source link establishes the error shape and context; the local snippet is the reproducible test input.

## Result

The current implementation produces **15 strict hits out of 20 selected failures**:

| Ecosystem | Strict hits | Cases |
| --- | ---: | ---: |
| Maven | 3 | 4 |
| Gradle | 3 | 4 |
| npm | 2 | 4 |
| pnpm | 3 | 4 |
| Docker | 4 | 4 |
| **Total** | **15** | **20** |

A hit requires the rendered diagnosis to use the expected tool family and contain the manually labelled root-cause fragment. A wrapper such as `code ERESOLVE`, `There are test failures`, or `Could not resolve all files` does not count when the selected input contains a more specific cause.

BuildSift currently renders pnpm failures under the `npm` tool family. The manifest records that current behavior explicitly, while keeping `pnpm` as the input ecosystem. Consequently, these numbers do not test whether npm and pnpm are labelled separately.

## Cases

| ID | Ecosystem | Expected root cause | Current result | Public source |
| --- | --- | --- | --- | --- |
| `maven-compiler-symbol` | Maven | Java compiler cannot find a referenced symbol | **Hit** — returns `cannot find symbol` | [MapStruct issue #1270](https://github.com/mapstruct/mapstruct/issues/1270) |
| `maven-parent-resolution` | Maven | Parent POM transfer times out | **Hit** — preserves `connect timed out` in the diagnosis | [Spring Boot issue #14970](https://github.com/spring-projects/spring-boot/issues/14970) |
| `maven-dependency-missing` | Maven | Requested JAXB artifact is absent from the configured repository | **Hit** — preserves the missing artifact coordinate | [Spring Boot issue #15102](https://github.com/spring-projects/spring-boot/issues/15102) |
| `maven-surefire-assertion` | Maven | Test assertion expected `4` but received `5` | **Miss** — returns the Surefire `There are test failures` wrapper | [Apache Surefire source test](https://maven.apache.org/surefire/maven-surefire-plugin/xref-test/org/apache/maven/plugin/surefire/SurefireMojoTest.html) |
| `gradle-removed-compile` | Gradle | Removed `compile` configuration is still used | **Hit** — returns `Could not find method compile()` | [Gradle 7 upgrade guide](https://docs.gradle.org/current/userguide/upgrading_version_6.html) |
| `gradle-jdk-major` | Gradle | Running a Gradle/Groovy combination that cannot read Java class version 66 | **Hit** — returns `Unsupported class file major version 66` | [Gradle issue #26162](https://github.com/gradle/gradle/issues/26162) |
| `gradle-config-cache` | Gradle | A task accesses `Task.project` at execution time with configuration cache enabled | **Hit** — returns the unsupported invocation | [Gradle configuration-cache troubleshooting](https://docs.gradle.org/current/userguide/configuration_cache_debugging.html) |
| `gradle-dependency-pkix` | Gradle | TLS certificate chain validation fails during dependency download | **Miss** — stops at `Could not resolve all files` instead of the deeper PKIX cause | [Gradle issue #32421](https://github.com/gradle/gradle/issues/32421) |
| `npm-missing-module` | npm | Runtime dependency `promise-retry` is missing | **Hit** — returns `Cannot find module 'promise-retry'` | [npm CLI issue #9151](https://github.com/npm/cli/issues/9151) |
| `npm-enoent-file` | npm | Required `package.json` file does not exist | **Hit** — returns `no such file or directory` | [npm CLI issue #3910](https://github.com/npm/cli/issues/3910) |
| `npm-peer-eresolve` | npm | Installed React version violates a package's peer range | **Miss** — returns only `code ERESOLVE` | [npm CLI issue #4998](https://github.com/npm/cli/issues/4998) |
| `npm-eacces-global` | npm | Global install directory is not writable | **Miss** — returns only `code EACCES` | [npm permissions guide](https://docs.npmjs.com/resolving-eacces-permissions-errors-when-installing-packages-globally/) |
| `pnpm-outdated-lockfile` | pnpm | Lockfile and workspace package manifest are out of sync | **Hit** — returns `ERR_PNPM_OUTDATED_LOCKFILE` and its explanation | [pnpm issue #7672](https://github.com/pnpm/pnpm/issues/7672) |
| `pnpm-peer-dependencies` | pnpm | Required peer dependencies are missing | **Hit** — returns `ERR_PNPM_PEER_DEP_ISSUES` | [pnpm issue #5152](https://github.com/pnpm/pnpm/issues/5152) |
| `pnpm-no-matching-version` | pnpm | Requested version is unavailable under the active release policy | **Hit** — returns `ERR_PNPM_NO_MATCHING_VERSION` | [pnpm issue #10014](https://github.com/pnpm/pnpm/issues/10014) |
| `pnpm-fetch-auth` | pnpm | Private registry request has no authorization header | **Miss** — returns the 404 wrapper but omits the authentication cause | [pnpm issue #5970](https://github.com/pnpm/pnpm/issues/5970) |
| `docker-daemon-unavailable` | Docker | Docker client cannot reach the daemon | **Hit** — returns the daemon connection error | [Docker for Linux issue #535](https://github.com/docker/for-linux/issues/535) |
| `docker-port-allocated` | Docker | Requested host port is already allocated | **Hit** — returns `port is already allocated` | [Docker Compose issue #4950](https://github.com/docker/compose/issues/4950) |
| `docker-no-manifest` | Docker | Image tag has no manifest for `linux/arm64/v8` | **Hit** — returns the missing-platform manifest error | [Docker Compose issue #9889](https://github.com/docker/compose/issues/9889) |
| `docker-missing-dockerfile` | Docker | The selected Dockerfile path does not exist | **Hit** — returns `no such file or directory` | [Docker Compose issue #7397](https://github.com/docker/compose/issues/7397) |

## Reproduce

From the repository root:

```sh
sh benchmark/run.sh
```

The script builds the current CLI into a temporary directory, runs every file in `benchmark/testdata`, prints each observed result, and removes its temporary Go caches and binary. It does not update the analyzer, fixtures, or baseline. It exits non-zero if observed hit/miss behavior drifts from `benchmark/cases.tsv`, so any intentional improvement requires a human to review and update the recorded baseline.

## What this does not establish

- The cases were selected after inspecting the current rules. They are suitable for validation and regression checks, not an unbiased accuracy estimate.
- The set contains only failures. It cannot measure false-positive rate or precision on successful, warning-only, or unrelated logs.
- Minimal snippets are easier than complete noisy logs and do not represent the distribution of tools, versions, languages, operating systems, or CI providers in the wild.
- A string hit confirms that the expected evidence reached the diagnosis. It does not prove that the proposed next command fixes the underlying build.
- Several public reports have context-specific causes. The labels here describe only the minimized local input, not every root cause discussed in the linked thread.

A defensible accuracy claim needs a larger, versioned, independently labelled, held-out corpus of complete sanitized logs, including negative examples and an adjudication process for ambiguous failures.
