# chart-perly-w3 chain export

`milestone.json` and `issues.json` are `git zhi milestone show chart-perly-w3 --format json` and the milestone's 12 issues from `git zhi issue list --all --format json`, exported on 2026-11-24 after completion.

They are the durable copy of a **reconstructed** chain. The original chain lived only in one container and was lost before it was ever synced to origin; it was rebuilt from the session transcript under perigrin's ruling "Partial rebuild, honest (Recommended)", and kept by "Accept + export JSON (Recommended)". The milestone's Records state what was recovered and what was lost. Issue ids are new; each issue body names the original id the delivered commits carry as `Implements:`.

The chain itself (`refs/zhi/*`) could not be pushed from the session that rebuilt it (the git proxy refused it with a policy 403); these files stand in until it is.
