package deploy

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	k8syaml "k8s.io/apimachinery/pkg/util/yaml"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

// getKubeConfig returns the Kubernetes REST config.
func getKubeConfig() (*rest.Config, error) {
	// Try in-cluster config first
	config, err := rest.InClusterConfig()
	if err == nil {
		return config, nil
	}

	// Fall back to kubeconfig file
	kubeconfig := os.Getenv("KUBECONFIG")
	if kubeconfig == "" {
		kubeconfig = clientcmd.NewDefaultClientConfigLoadingRules().GetDefaultFilename()
	}

	config, err = clientcmd.BuildConfigFromFlags("", kubeconfig)
	if err != nil {
		return nil, fmt.Errorf("failed to build kubeconfig: %w", err)
	}

	return config, nil
}

// getKubernetesClient returns a Kubernetes clientset.
func getKubernetesClient() (*kubernetes.Clientset, error) {
	config, err := getKubeConfig()
	if err != nil {
		return nil, err
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create kubernetes client: %w", err)
	}

	return clientset, nil
}

// getDynamicClient returns a dynamic Kubernetes client.
func getDynamicClient() (dynamic.Interface, error) {
	config, err := getKubeConfig()
	if err != nil {
		return nil, err
	}

	client, err := dynamic.NewForConfig(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create dynamic client: %w", err)
	}

	return client, nil
}

// ApplyResources applies the generated YAML resources to the Kubernetes cluster using the Kubernetes SDK.
func ApplyResources(ctx context.Context, projectRoot string, namespace string, dryRun bool) error {
	deployDir := GetDeployDir(projectRoot)

	// Check if deploy directory exists
	if _, err := os.Stat(deployDir); os.IsNotExist(err) {
		return fmt.Errorf("deploy directory does not exist: %s", deployDir)
	}

	// Get dynamic client for applying resources
	dynamicClient, err := getDynamicClient()
	if err != nil {
		return fmt.Errorf("get kubernetes client: %w", err)
	}

	// Resources to apply in order
	// DopplerSecret must be applied first so the secret exists when Knative Service references it
	resources := []struct {
		name string
		path string
	}{
		{"DopplerSecret", GetDopplerSecretPath(projectRoot)},
		{"Knative Service", GetKnativeServicePath(projectRoot)},
		{"Pipe", GetPipePath(projectRoot)},
	}

	for _, resource := range resources {
		// Check if file exists
		if _, err := os.Stat(resource.path); os.IsNotExist(err) {
			// Skip if file doesn't exist (e.g., integration might be optional)
			continue
		}

		if err := applyResourceFromFile(ctx, dynamicClient, resource.path, namespace, dryRun); err != nil {
			return fmt.Errorf("apply %s: %w", resource.name, err)
		}
	}

	return nil
}

// applyResourceFromFile applies a single YAML file using the Kubernetes SDK.
func applyResourceFromFile(ctx context.Context, dynamicClient dynamic.Interface, filePath, namespace string, dryRun bool) error {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("read file: %w", err)
	}

	// Parse YAML into unstructured object
	decoder := k8syaml.NewYAMLOrJSONDecoder(strings.NewReader(string(data)), 4096)
	obj := &unstructured.Unstructured{}
	if err := decoder.Decode(obj); err != nil {
		return fmt.Errorf("decode yaml: %w", err)
	}

	// Override namespace if provided
	if namespace != "" {
		obj.SetNamespace(namespace)
	}

	if dryRun {
		// For dry-run, just validate the object
		fmt.Fprintf(os.Stderr, "Dry-run: would apply %s/%s in namespace %s\n", obj.GetKind(), obj.GetName(), obj.GetNamespace())
		return nil
	}

	// Get GVR from the object using API discovery
	gvr, err := getGVRForObject(ctx, obj)
	if err != nil {
		return fmt.Errorf("get GVR for object: %w", err)
	}

	// Get namespace resource interface
	var resourceInterface dynamic.ResourceInterface
	if obj.GetNamespace() != "" {
		resourceInterface = dynamicClient.Resource(gvr).Namespace(obj.GetNamespace())
	} else {
		resourceInterface = dynamicClient.Resource(gvr)
	}

	// Apply using server-side apply
	_, err = resourceInterface.Apply(ctx, obj.GetName(), obj, metav1.ApplyOptions{
		FieldManager: "stax-cli",
		Force:        true,
	})
	if err != nil {
		return fmt.Errorf("apply resource: %w", err)
	}

	fmt.Fprintf(os.Stderr, "Applied %s/%s in namespace %s\n", obj.GetKind(), obj.GetName(), obj.GetNamespace())
	return nil
}

// getGVRForObject gets the GroupVersionResource for an object using API discovery.
func getGVRForObject(ctx context.Context, obj *unstructured.Unstructured) (schema.GroupVersionResource, error) {
	gvk := obj.GroupVersionKind()

	// Try to discover the resource using API discovery
	config, err := getKubeConfig()
	if err != nil {
		return schema.GroupVersionResource{}, err
	}

	discoveryClient, err := discovery.NewDiscoveryClientForConfig(config)
	if err != nil {
		return schema.GroupVersionResource{}, fmt.Errorf("create discovery client: %w", err)
	}

	// Get API resources for the group version
	apiResources, err := discoveryClient.ServerResourcesForGroupVersion(gvk.GroupVersion().String())
	if err != nil {
		// If discovery fails, fall back to heuristic mapping
		return schema.GroupVersionResource{
			Group:    gvk.Group,
			Version:  gvk.Version,
			Resource: getResourceFromKind(gvk.Kind),
		}, nil
	}

	// Find the resource that matches the kind
	for _, apiResource := range apiResources.APIResources {
		if apiResource.Kind == gvk.Kind {
			return schema.GroupVersionResource{
				Group:    gvk.Group,
				Version:  gvk.Version,
				Resource: apiResource.Name,
			}, nil
		}
	}

	// Fallback to heuristic if not found
	return schema.GroupVersionResource{
		Group:    gvk.Group,
		Version:  gvk.Version,
		Resource: getResourceFromKind(gvk.Kind),
	}, nil
}

// getResourceFromKind converts a Kind to a resource name (plural, lowercase).
// This is a fallback when API discovery fails.
func getResourceFromKind(kind string) string {
	// Mapping for resources we use (exact resource names)
	kindToResource := map[string]string{
		"DopplerSecret": "dopplersecrets",
		"Service":       "services",
		"Pipe":          "pipes",
	}

	if resource, ok := kindToResource[kind]; ok {
		return resource
	}

	// Fallback: convert to lowercase and pluralize (simple heuristic)
	lower := strings.ToLower(kind)
	if strings.HasSuffix(lower, "y") {
		return lower[:len(lower)-1] + "ies"
	}
	if strings.HasSuffix(lower, "s") || strings.HasSuffix(lower, "x") || strings.HasSuffix(lower, "z") {
		return lower + "es"
	}
	return lower + "s"
}

// CheckKubectl checks if Kubernetes cluster is accessible.
func CheckKubectl() error {
	_, err := getKubeConfig()
	if err != nil {
		return fmt.Errorf("kubernetes cluster not accessible: %w", err)
	}
	return nil
}

// EnsureDopplerTokenSecret ensures the service-specific doppler token secret exists in doppler-operator-system namespace.
// The secret name follows the pattern: {service_name}-token-secret
// If it doesn't exist, it will prompt for the token or use DOPPLER_SERVICE_TOKEN env var.
func EnsureDopplerTokenSecret(ctx context.Context, serviceName string) error {
	secretName := fmt.Sprintf("%s-token-secret", serviceName)
	namespace := "doppler-operator-system"

	// Check if secret already exists
	if secretExists(ctx, secretName, namespace) {
		fmt.Fprintf(os.Stderr, "Doppler token secret %s already exists in namespace %s\n", secretName, namespace)
		return nil
	}

	fmt.Fprintf(os.Stderr, "Doppler token secret %s does not exist in namespace %s\n", secretName, namespace)

	// Get the service token
	token, err := getDopplerServiceToken(serviceName)
	if err != nil {
		return fmt.Errorf("get doppler service token: %w", err)
	}

	if token == "" {
		return fmt.Errorf("doppler service token cannot be empty")
	}

	// Create the secret
	if err := createDopplerTokenSecret(ctx, secretName, namespace, token); err != nil {
		return fmt.Errorf("create doppler token secret: %w", err)
	}

	fmt.Fprintf(os.Stderr, "Created %s in namespace %s\n", secretName, namespace)
	return nil
}

// secretExists checks if a secret exists in the given namespace.
func secretExists(ctx context.Context, name, namespace string) bool {
	clientset, err := getKubernetesClient()
	if err != nil {
		return false
	}

	_, err = clientset.CoreV1().Secrets(namespace).Get(ctx, name, metav1.GetOptions{})
	return err == nil
}

// getDopplerServiceToken gets the Doppler service token from service-specific env var or prompts the user.
func getDopplerServiceToken(serviceName string) (string, error) {
	// First, try service-specific environment variable: DOPPLER_SERVICE_TOKEN_{SERVICE_NAME}
	// Convert service name to uppercase and replace hyphens with underscores for env var name
	envVarName := fmt.Sprintf("DOPPLER_SERVICE_TOKEN_%s", strings.ToUpper(strings.ReplaceAll(serviceName, "-", "_")))
	if token := os.Getenv(envVarName); token != "" {
		fmt.Fprintf(os.Stderr, "Using Doppler service token from %s environment variable\n", envVarName)
		return token, nil
	}

	// Fallback to generic DOPPLER_SERVICE_TOKEN for backward compatibility
	if token := os.Getenv("DOPPLER_SERVICE_TOKEN"); token != "" {
		fmt.Fprintf(os.Stderr, "Using Doppler service token from DOPPLER_SERVICE_TOKEN environment variable\n")
		return token, nil
	}

	// Prompt the user if environment variable is not set
	fmt.Fprintf(os.Stderr, "Doppler service token not found in %s or DOPPLER_SERVICE_TOKEN environment variable.\n", envVarName)
	return promptForDopplerToken(serviceName)
}

// promptForDopplerToken prompts the user for the Doppler service token.
func promptForDopplerToken(serviceName string) (string, error) {
	fmt.Fprintf(os.Stderr, "Please enter your Doppler service token for service '%s': ", serviceName)

	scanner := bufio.NewScanner(os.Stdin)
	if !scanner.Scan() {
		return "", fmt.Errorf("failed to read input: %w", scanner.Err())
	}

	token := strings.TrimSpace(scanner.Text())
	if token == "" {
		return "", fmt.Errorf("doppler service token cannot be empty")
	}

	return token, nil
}

// createDopplerTokenSecret creates the doppler-token-secret using the Kubernetes SDK.
func createDopplerTokenSecret(ctx context.Context, name, namespace, token string) error {
	clientset, err := getKubernetesClient()
	if err != nil {
		return fmt.Errorf("get kubernetes client: %w", err)
	}

	secret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Type: corev1.SecretTypeOpaque,
		StringData: map[string]string{
			"serviceToken": token,
		},
	}

	_, err = clientset.CoreV1().Secrets(namespace).Create(ctx, secret, metav1.CreateOptions{})
	if err != nil {
		if errors.IsAlreadyExists(err) {
			// Secret already exists, update it
			existing, getErr := clientset.CoreV1().Secrets(namespace).Get(ctx, name, metav1.GetOptions{})
			if getErr != nil {
				return fmt.Errorf("get existing secret: %w", getErr)
			}
			existing.StringData = map[string]string{
				"serviceToken": token,
			}
			_, err = clientset.CoreV1().Secrets(namespace).Update(ctx, existing, metav1.UpdateOptions{})
			if err != nil {
				return fmt.Errorf("update secret: %w", err)
			}
			return nil
		}
		return fmt.Errorf("create secret: %w", err)
	}

	return nil
}
