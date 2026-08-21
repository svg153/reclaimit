# Growth measurement and privacy

reclaimit should grow through useful, verifiable material: a reproducible demo,
benchmarks with methodology, precise safety documentation, and category work
chosen from user evidence. Unsupported recovery totals, generic speed claims,
and unverified competitor comparisons are excluded.

The search position is deliberately specific: **developer disk cleanup CLI**
for regenerable artifacts, used alongside general disk analyzers. README and
landing copy may name ncdu, gdu, dust, and dua for that complementary workflow,
but must not claim they lack features without checking their current official
documentation. “Alternative” queries should explain the product boundary, not
pretend reclaimit is a drop-in replacement.

Do not publish the supplied marketing report's “30–50 GB”, “47 GB”, Homebrew,
Docker-layer, Xcode-cache, plugin-system, or expected-traffic claims. They are
not supported by the current code, distribution channels, or measured data.

## Baseline

The initial audit snapshot was taken on 2026-08-14. It is a historical baseline and
must not be read as the current release state.

| Signal | Baseline | Source or status |
| --- | ---: | --- |
| GitHub stars | 3 | Public repository metadata captured in `docs/code-review.md` |
| GitHub forks | 1 | Public repository metadata captured in `docs/code-review.md` |
| Latest-release asset downloads | 1 | Sum captured during the repository audit |
| GitHub views and unique visitors | Not captured | Owner-only Traffic view was unavailable to this workspace |
| GitHub clones and unique cloners | Not captured | Owner-only Traffic view was unavailable to this workspace |
| Pages search impressions | Not captured | No Search Console property was available |

“Not captured” is deliberate missing data, not zero. Record those values from
GitHub Insights → Traffic and the configured search-console property on the day
the next release is published. Store only aggregate counts.

## Release cohort

For each release, record these values on publication day and again after 30
days:

| Release | Published UTC | Stars at publish | Stars +30d | Downloads +30d | Views +30d | Unique clones +30d | Search impressions +30d |
| --- | --- | ---: | ---: | ---: | ---: | ---: | ---: |
| `v0.4.2` | 2026-08-21 | 3 | Pending | Pending | Pending | Pending | Pending |

Use differences, not absolute totals, when attributing a release cohort. Treat
correlation as directional evidence rather than proof that one post caused a
star or download.

## Thirty-day decision rule

After 30 days, choose one next investment:

1. expand demo/release content when qualified repository visits and clones
   increased;
2. improve install and onboarding when visits increased but clones or downloads
   did not;
3. implement the best-specified category request when opt-in community evidence
   identifies a repeated gap;
4. revise search copy when Pages impressions increased but click-through did
   not.

The opt-in category thread is [issue #54](https://github.com/svg153/reclaimit/issues/54).
It asks for placeholders and rounded category totals, not complete reports.

## Data boundary

reclaimit has no telemetry and sends no scan data by default. Do not add network
collection to support growth reporting. Any research dataset must be explicitly
opt-in, aggregate-only, documented before collection, and safe to delete on
request. Never collect absolute paths, repository names, filenames, file
contents, usernames, credentials, or customer data.
