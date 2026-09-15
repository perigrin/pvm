// ABOUTME: Rejects GitHub contexts used where they do not exist, which makes a workflow unparseable.
// ABOUTME: Three real runs died 0s with "workflow file issue" and every local guard stayed green.

package parseoracle_test

import (
	"os"
	"strings"
	"testing"

	"gopkg.in/yaml.v3"
)

// TestWorkflowContextsAreAvailableWhereUsed catches the failure that no other
// guard in this package could see.
//
// `runner.temp` is not available in a job-level `env:`. A workflow that names
// it there does not parse: GitHub records a 0-second failure, "This run
// likely failed because of a workflow file issue", creates no jobs, and runs
// nothing. Measured on runs 34791954231, 34808903487 and 34848323354 --
// including one predating every other change to that file, which is what
// ruled the other edits out.
//
// Every local guard stayed green through all three, because they read the
// file as text and as YAML and it is valid as both. The defect is in what
// GitHub will EVALUATE, not in what the YAML says, so it needs its own rule.
//
// The contexts available at job level are a documented subset:
// github, needs, vars, inputs, always the `env` of the workflow. `runner`,
// `steps`, `job`, `matrix` and `secrets` are not among them, and `runner` is
// the one this file actually reaches for -- every path it builds is under
// $RUNNER_TEMP.
func TestWorkflowContextsAreAvailableWhereUsed(t *testing.T) {
	// Contexts that do not exist in a job-level env: block. A job-level env is
	// evaluated before any runner is assigned, which is why `runner` is out.
	forbidden := []string{"runner.", "steps.", "job.", "matrix."}

	data, err := os.ReadFile(repoFile(t, ".github", "workflows", workflow))
	if err != nil {
		t.Fatalf("reading the workflow: %v", err)
	}

	var wf struct {
		Jobs map[string]struct {
			Env map[string]string `yaml:"env"`
		} `yaml:"jobs"`
	}
	if err := yaml.Unmarshal(data, &wf); err != nil {
		t.Fatalf("parsing the workflow: %v", err)
	}

	for name, job := range wf.Jobs {
		for key, value := range job.Env {
			for _, ctx := range forbidden {
				if strings.Contains(value, ctx) {
					t.Errorf("job %q sets env %s to %q, which names the %s context: "+
						"that context does not exist in a job-level env, so GitHub "+
						"refuses the whole workflow and runs nothing. Set it on the "+
						"steps that need it instead.",
						name, key, value, strings.TrimSuffix(ctx, "."))
				}
			}
		}
	}
}
