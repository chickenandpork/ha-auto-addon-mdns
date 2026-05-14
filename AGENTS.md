## Constraints & Hints

- Use Bazel for all builds and tests.
- If `bazel query //...` fails, fix it locally before proceeding — it must succeed.
- Run `bazel test //...` before committing; this is a requirement.
- Make small, iterative changes and commit with passing tests.
- PR once the feature is complete.
- don't go searching for existing caches of resources:
  - downloaded artifacts that are used in MODULE.bazel can be stored in ${HOME}/.cache/distdir where bazel will automatically look for them.
  - set a distdir=${HOME}/.cache/distdir in .bazelrc to tell bazel to look in that directory
  - set a repository_cache=${HOME}/.cache/bazel-repo-cache in .bazelrc to tell bazel to look in that directory
  - set a disk_cache=${HOME}/.cache/bazel-disk-cache in .bazelrc to tell bazel to look in that directory
  - if you cannot access the directories for distdir, repository_cache, disk_cache, request permission

## Bazel constraints
- use Bazel Central Repo for all bazel dependencies 
- don't use WORKSPACE files; use the MODULE.bazel and BUILD.bazel files
- if there's a missing go.sum value, use "bazel run @rules_go//go -- mod tidy"
- don't use "io_bazel_rules_go" or "bazel_gazelle" as resource names; use the standard BCR values "rules_go" and "gazelle"
- "bazel" is always run via "/usr/local/bin/subazel"
- add to, or edit, but don't delete, .bazelrc and .bazelversion
- do not edit files in /Users/allanc/Library/Caches/ as long-term fixes, but only to explore patches for downloaded dependencies
  - these files are ephemeral, and will be destroyed on rebuilds
  - ultimately, patches appled to dependencies on download is the appropriate means to edit dependencies, if at all
- DO NOT use `go_sdk.host(...)` in MODULE.bazel; rather, use `go_sdk.from_file("//:go,mod")` to pick up the version from the go.mod file, keeping a single place to define the version used

## Go constraints
- never directly run the "go" tool.  Always run "go" through "bazel run @rules_go//go -- " for example: "go mod tidy"
  would be "bazel run @rules_go//go -- mod tidy".
- Do not go digging around the filesystem for random scraps of go dependencies
  - set go versions in go.mod
  - set go dependencies in go.mod
  - update go.sum using "bazel run @rules_go//go -- mod tidy"
- after any sort of changes to the go.mod or go.sum files, run "bazel mod tidy"
- if a dependency is present as an indirect dependency in go.mod, but shows as needed during a build, change that dependency to a direct dependency.  For example, if "com_github_example_bluey" is missing as a dependency, but "github.com/example/bluey" exists as a dependency in go.mod but commented "indirect", remove that "indirect" on that dependency, re-run "bazel run @rules_go//go -- mod tidy", then if "com_github_example_bluey" isn't listed in a "use_repos" line in MODULE.bazel, re-run "bazel mod tidy"
- add to, but don't delete, go.sum and go.mod.
- the go toolchain version is written in go.mod; all other needs must read from that single place.  the "from_file" tag from the extensions in rules_go and gazelle are recommended means of doing so.


## Git constraints
- add to, but don't delete, .gitignore
- all work is in a git branch.  Make a new branch if you need to.  completed work is checked into that branch.
- follow-on work should either stack another commit, or if changing to the next step, a new branch based from the current work.


## Roadmap and Constraints
- Add tasks considered less-critical to a section in ROADMAP.md titled "deferred todo items"
- Any other rules, add to the appropriate section in AGENTS.md.  If it's unclear what section to add to, prompt for a section name to create or use.
- add to, edit, but don't delete these ROADMAP.md and AGENTS.md
