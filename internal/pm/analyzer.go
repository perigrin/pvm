// ABOUTME: Module analyzer stub for PM
// ABOUTME: Returns "not yet available" since analysis requires type-system components

package pm

import (
	"fmt"
)

// ModuleAnalyzer is a stub for the real module analyzer
type ModuleAnalyzer struct{}

// NewModuleAnalyzer creates a stub module analyzer
func NewModuleAnalyzer() (*ModuleAnalyzer, error) {
	return &ModuleAnalyzer{}, nil
}

// AnalyzeModule is a stub that returns an error indicating the analyzer is not yet available
func (ma *ModuleAnalyzer) AnalyzeModule(modulePath string) (interface{}, error) {
	return nil, fmt.Errorf("module analysis is not yet available in this build")
}
