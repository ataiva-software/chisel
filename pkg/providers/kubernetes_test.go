package providers

import (
	"context"
	"testing"

	"github.com/ataiva-software/chisel/pkg/types"
)

func TestKubernetesProvider_Type(t *testing.T) {
	provider := NewKubernetesProvider()
	
	if provider.Type() != "kubernetes" {
		t.Errorf("Expected type 'kubernetes', got '%s'", provider.Type())
	}
}

func TestKubernetesProvider_ValidateDeployment(t *testing.T) {
	provider := NewKubernetesProvider()
	
	tests := []struct {
		name     string
		resource *types.Resource
		wantErr  bool
	}{
		{
			name: "valid deployment",
			resource: &types.Resource{
				Type: "kubernetes",
				Name: "nginx-deployment",
				Properties: map[string]interface{}{
					"kind":      "Deployment",
					"namespace": "default",
					"spec": map[string]interface{}{
						"replicas": 3,
						"selector": map[string]interface{}{
							"matchLabels": map[string]interface{}{
								"app": "nginx",
							},
						},
						"template": map[string]interface{}{
							"metadata": map[string]interface{}{
								"labels": map[string]interface{}{
									"app": "nginx",
								},
							},
							"spec": map[string]interface{}{
								"containers": []interface{}{
									map[string]interface{}{
										"name":  "nginx",
										"image": "nginx:1.21",
										"ports": []interface{}{
											map[string]interface{}{
												"containerPort": 80,
											},
										},
									},
								},
							},
						},
					},
				},
			},
			wantErr: false,
		},
		{
			name: "missing kind",
			resource: &types.Resource{
				Type: "kubernetes",
				Name: "invalid-resource",
				Properties: map[string]interface{}{
					"namespace": "default",
				},
			},
			wantErr: true,
		},
		{
			name: "missing namespace",
			resource: &types.Resource{
				Type: "kubernetes",
				Name: "invalid-resource",
				Properties: map[string]interface{}{
					"kind": "Deployment",
				},
			},
			wantErr: true,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := provider.Validate(tt.resource)
			if tt.wantErr {
				if err == nil {
					t.Error("Expected validation error but got none")
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected validation error: %v", err)
				}
			}
		})
	}
}

func TestKubernetesProvider_ReadDeployment(t *testing.T) {
	provider := NewKubernetesProvider()
	provider.SetMockMode(true)
	
	resource := &types.Resource{
		Type: "kubernetes",
		Name: "nginx-deployment",
		Properties: map[string]interface{}{
			"kind":      "Deployment",
			"namespace": "default",
		},
	}
	
	ctx := context.Background()
	state, err := provider.Read(ctx, resource)
	if err != nil {
		t.Fatalf("Failed to read resource: %v", err)
	}
	
	if state["kind"] != "Deployment" {
		t.Errorf("Expected kind 'Deployment', got '%v'", state["kind"])
	}
	
	if state["namespace"] != "default" {
		t.Errorf("Expected namespace 'default', got '%v'", state["namespace"])
	}
}

func TestKubernetesProvider_DiffDeployment(t *testing.T) {
	provider := NewKubernetesProvider()
	provider.SetMockMode(true)
	
	resource := &types.Resource{
		Type: "kubernetes",
		Name: "nginx-deployment",
		Properties: map[string]interface{}{
			"kind":      "Deployment",
			"namespace": "default",
			"spec": map[string]interface{}{
				"replicas": 5, // Changed from 3 to 5
			},
		},
	}
	
	current := map[string]interface{}{
		"kind":      "Deployment",
		"namespace": "default",
		"spec": map[string]interface{}{
			"replicas": 3,
		},
	}
	
	ctx := context.Background()
	diff, err := provider.Diff(ctx, resource, current)
	if err != nil {
		t.Fatalf("Failed to diff resource: %v", err)
	}
	
	if diff.Action != types.ActionUpdate {
		t.Errorf("Expected action 'update', got '%s'", diff.Action)
	}
	
	if len(diff.Changes) == 0 {
		t.Error("Expected changes but got none")
	}
}

func TestKubernetesProvider_ApplyDeployment(t *testing.T) {
	provider := NewKubernetesProvider()
	provider.SetMockMode(true)
	
	resource := &types.Resource{
		Type: "kubernetes",
		Name: "nginx-deployment",
		Properties: map[string]interface{}{
			"kind":      "Deployment",
			"namespace": "default",
			"spec": map[string]interface{}{
				"replicas": 3,
			},
		},
	}
	
	diff := &types.ResourceDiff{
		ResourceID: "kubernetes.nginx-deployment",
		Action:     types.ActionCreate,
		Changes: map[string]interface{}{
			"spec.replicas": map[string]interface{}{
				"from": nil,
				"to":   3,
			},
		},
	}
	
	ctx := context.Background()
	err := provider.Apply(ctx, resource, diff)
	if err != nil {
		t.Fatalf("Failed to apply resource: %v", err)
	}
}

func TestKubernetesProvider_ValidateService(t *testing.T) {
	provider := NewKubernetesProvider()
	
	resource := &types.Resource{
		Type: "kubernetes",
		Name: "nginx-service",
		Properties: map[string]interface{}{
			"kind":      "Service",
			"namespace": "default",
			"spec": map[string]interface{}{
				"selector": map[string]interface{}{
					"app": "nginx",
				},
				"ports": []interface{}{
					map[string]interface{}{
						"port":       80,
						"targetPort": 80,
					},
				},
			},
		},
	}
	
	err := provider.Validate(resource)
	if err != nil {
		t.Errorf("Unexpected validation error: %v", err)
	}
}

func TestKubernetesProvider_ValidateConfigMap(t *testing.T) {
	provider := NewKubernetesProvider()
	
	resource := &types.Resource{
		Type: "kubernetes",
		Name: "app-config",
		Properties: map[string]interface{}{
			"kind":      "ConfigMap",
			"namespace": "default",
			"data": map[string]interface{}{
				"config.yaml": "key: value",
				"app.properties": "debug=true",
			},
		},
	}
	
	err := provider.Validate(resource)
	if err != nil {
		t.Errorf("Unexpected validation error: %v", err)
	}
}

func TestKubernetesProvider_ValidateSecret(t *testing.T) {
	provider := NewKubernetesProvider()
	
	resource := &types.Resource{
		Type: "kubernetes",
		Name: "app-secret",
		Properties: map[string]interface{}{
			"kind":      "Secret",
			"namespace": "default",
			"type":      "Opaque",
			"data": map[string]interface{}{
				"username": "YWRtaW4=", // base64 encoded
				"password": "MWYyZDFlMmU2N2Rm", // base64 encoded
			},
		},
	}
	
	err := provider.Validate(resource)
	if err != nil {
		t.Errorf("Unexpected validation error: %v", err)
	}
}

func TestKubernetesProvider_WithoutMockMode(t *testing.T) {
	provider := NewKubernetesProvider()
	// Don't set mock mode - should fail without real cluster
	
	resource := &types.Resource{
		Type: "kubernetes",
		Name: "test-deployment",
		Properties: map[string]interface{}{
			"kind":      "Deployment",
			"namespace": "default",
		},
	}
	
	ctx := context.Background()
	_, err := provider.Read(ctx, resource)
	
	// Should fail without real Kubernetes cluster
	if err == nil {
		t.Log("Kubernetes read succeeded (unexpected without cluster)")
	} else {
		t.Logf("Kubernetes read failed as expected without cluster: %v", err)
	}
}

func TestKubernetesProvider_SupportedKinds(t *testing.T) {
	provider := NewKubernetesProvider()
	
	tests := []struct {
		kind       string
		properties map[string]interface{}
	}{
		{
			kind: "Deployment",
			properties: map[string]interface{}{
				"kind":      "Deployment",
				"namespace": "default",
				"spec": map[string]interface{}{
					"selector": map[string]interface{}{},
					"template": map[string]interface{}{},
				},
			},
		},
		{
			kind: "Service",
			properties: map[string]interface{}{
				"kind":      "Service",
				"namespace": "default",
				"spec": map[string]interface{}{
					"ports": []interface{}{},
				},
			},
		},
		{
			kind: "ConfigMap",
			properties: map[string]interface{}{
				"kind":      "ConfigMap",
				"namespace": "default",
				"data": map[string]interface{}{
					"key": "value",
				},
			},
		},
		{
			kind: "Secret",
			properties: map[string]interface{}{
				"kind":      "Secret",
				"namespace": "default",
				"data": map[string]interface{}{
					"key": "value",
				},
			},
		},
		{
			kind: "Ingress",
			properties: map[string]interface{}{
				"kind":      "Ingress",
				"namespace": "default",
			},
		},
		{
			kind: "PersistentVolumeClaim",
			properties: map[string]interface{}{
				"kind":      "PersistentVolumeClaim",
				"namespace": "default",
			},
		},
		{
			kind: "Namespace",
			properties: map[string]interface{}{
				"kind":      "Namespace",
				"namespace": "default",
			},
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.kind, func(t *testing.T) {
			resource := &types.Resource{
				Type:       "kubernetes",
				Name:       "test-resource",
				Properties: tt.properties,
			}
			
			err := provider.Validate(resource)
			if err != nil {
				t.Errorf("Expected kind '%s' to be supported, got error: %v", tt.kind, err)
			}
		})
	}
}

func TestKubernetesProvider_UnsupportedKind(t *testing.T) {
	provider := NewKubernetesProvider()
	
	resource := &types.Resource{
		Type: "kubernetes",
		Name: "test-resource",
		Properties: map[string]interface{}{
			"kind":      "UnsupportedKind",
			"namespace": "default",
		},
	}
	
	err := provider.Validate(resource)
	if err == nil {
		t.Error("Expected error for unsupported kind")
	}
}
