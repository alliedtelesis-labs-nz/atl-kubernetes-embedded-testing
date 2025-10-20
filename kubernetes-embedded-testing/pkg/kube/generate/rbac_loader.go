package generate

import (
	"fmt"
	"os"

	rbacv1 "k8s.io/api/rbac/v1"
	"sigs.k8s.io/yaml"
)

// RBACRulesFile represents the structure of an RBAC rules YAML file
type RBACRulesFile struct {
	Rules []rbacv1.PolicyRule `yaml:"rules" json:"rules"`
}

// LoadRBACRulesFromFile loads additional RBAC rules from a YAML file
func LoadRBACRulesFromFile(filepath string) ([]rbacv1.PolicyRule, error) {
	if filepath == "" {
		return nil, nil
	}

	data, err := os.ReadFile(filepath)
	if err != nil {
		return nil, fmt.Errorf("failed to read RBAC file %s: %w", filepath, err)
	}

	var rbacFile RBACRulesFile
	if err := yaml.Unmarshal(data, &rbacFile); err != nil {
		return nil, fmt.Errorf("failed to parse RBAC file %s: %w", filepath, err)
	}

	return rbacFile.Rules, nil
}

// MergeRBACRules merges default RBAC rules with additional rules from a file
func MergeRBACRules(defaultRules []rbacv1.PolicyRule, additionalRules []rbacv1.PolicyRule) []rbacv1.PolicyRule {
	if len(additionalRules) == 0 {
		return defaultRules
	}

	merged := make([]rbacv1.PolicyRule, len(defaultRules))
	copy(merged, defaultRules)
	merged = append(merged, additionalRules...)
	
	return merged
}
