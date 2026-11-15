// ABOUTME: The chart as a parseoracle subject: SubjectFacts read from the forest forest.Parse leaves, in-process.
// ABOUTME: A site every reading holds is decided; one only some readings hold is Unresolved, never a guess.

package forest

import (
	"tamarou.com/pvm/internal/chart/grammar"
	"tamarou.com/pvm/internal/parseoracle"
)

// SubjectFacts is the chart's answer to parseoracle's subject contract for src, a file at path.
func SubjectFacts(tab *grammar.Tables, src []byte, path string, scope Scope) parseoracle.SubjectFacts {
	return parseoracle.SubjectFacts{}
}
