# GitHub Action

BuildSift includes a composite action that analyzes a saved build log without sending it to an external service. GitHub-hosted Linux, macOS, and Windows runners include the Go toolchain required by the action.

```yaml
- name: Analyze failed build
  id: buildsift
  if: always()
  uses: topchen2025/buildsift@v0.1.0
  with:
    log-path: build.log

- name: Consume the evidence package
  if: always()
  shell: bash
  env:
    BUILDSIFT_JSON: ${{ steps.buildsift.outputs.json }}
  run: printf '%s\n' "$BUILDSIFT_JSON"
```

`log-path` may be absolute or relative to `GITHUB_WORKSPACE`. The `json` output contains the same structured evidence package printed by `buildsift --json`. The action also appends that package to the job summary.

```json
{
  "schema_version": "1",
  "tool": "maven",
  "confidence": "high",
  "root_cause": "NoSuchFileException: ~/work/quality/target/pmd.xml",
  "evidence": [
    {"line": 1842, "text": "NoSuchFileException: ~/work/quality/target/pmd.xml"}
  ],
  "cascade_count": 17,
  "next_check": "mvn -e -X",
  "version": "0.1.0"
}
```

When no supported pattern is found, `root_cause` is empty, `evidence` is an empty array, and `confidence` is `low`. Consumers should check those fields instead of assuming every log has a diagnosis.

Replace `v0.1.0` with a newer release tag when upgrading.
