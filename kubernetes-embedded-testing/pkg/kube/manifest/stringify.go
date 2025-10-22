package manifest

import (
	"testrunner/pkg/config"
	"testrunner/pkg/kube/generate"

	rbacv1 "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer/json"
	"k8s.io/client-go/kubernetes/scheme"
)

// marshalKubernetesObject properly marshals a Kubernetes object with TypeMeta fields
func marshalKubernetesObject(obj runtime.Object) ([]byte, error) {
	serializer := json.NewSerializerWithOptions(
		json.DefaultMetaFactory,
		scheme.Scheme,
		scheme.Scheme,
		json.SerializerOptions{
			Yaml:   true,
			Pretty: true,
			Strict: false,
		},
	)

	return runtime.Encode(serializer, obj)
}

// All generates all manifests as YAML strings
func All(cfg config.Config, namespace string) ([]string, error) {
	ns := generate.Namespace(namespace)
	
	// Load additional RBAC rules from file if specified
	var additionalRules []rbacv1.PolicyRule
	if cfg.RbacFile != "" {
		rules, err := generate.LoadRBACRulesFromFile(cfg.RbacFile)
		if err != nil {
			return nil, err
		}
		additionalRules = rules
	}
	
	role := generate.Role(namespace, additionalRules...)
	roleBinding := generate.RoleBinding(namespace)
	job, err := generate.Job(cfg, namespace)
	if err != nil {
		return nil, err
	}

	manifests := []runtime.Object{ns, sa, role, roleBinding, job}
	results := make([]string, len(manifests))

	for i, manifest := range manifests {
		yamlData, err := marshalKubernetesObject(manifest)
		if err != nil {
			return nil, err
		}
		results[i] = string(yamlData)
	}

	return results, nil
}
