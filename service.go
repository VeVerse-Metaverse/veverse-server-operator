package main

import (
	"context"
	"fmt"
	"github.com/gofrs/uuid"
	apiV1 "k8s.io/api/core/v1"
	metaV1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

func getGameServerServiceClusterResource(ctx context.Context, id uuid.UUID) (*apiV1.Service, error) {
	clientset := ctx.Value("clientset").(*kubernetes.Clientset)
	namespace := ctx.Value("namespace").(string)

	resourceName := getResourceName(id)

	serviceClient := clientset.CoreV1().Services(namespace)

	service, err := serviceClient.Get(ctx, resourceName, metaV1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to get service: %v", err)
	}

	return service, nil
}

func createGameServerServiceClusterResource(ctx context.Context, metadata unstructured.Unstructured) error {
	namespace, ok := ctx.Value("namespace").(string)
	if !ok {
		return fmt.Errorf("namespace not found in context")
	}

	config, ok := ctx.Value("config").(*rest.Config)
	if !ok {
		return fmt.Errorf("config not found in context")
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return fmt.Errorf("failed to create a clientset: %v", err)
	}

	//region Specification

	spec, ok := metadata.Object["spec"].(map[string]interface{})
	if !ok {
		return fmt.Errorf("spec not found in game server metadata")
	}

	//region ID
	if _, ok = spec["id"].(string); !ok {
		return fmt.Errorf("id not found in game server metadata")
	}

	id, err := uuid.FromString(spec["id"].(string))
	if err != nil {
		return fmt.Errorf("failed to parse id: %v", err)
	}

	resourceName := getResourceName(id)
	//endregion

	//endregion

	serviceClient := clientset.CoreV1().Services(namespace)
	serviceResource := &apiV1.Service{
		ObjectMeta: metaV1.ObjectMeta{
			Name: resourceName,
			Labels: map[string]string{
				"app": resourceName,
			},
		},
		Spec: apiV1.ServiceSpec{
			Selector: map[string]string{
				"app": resourceName,
			},
			Ports: []apiV1.ServicePort{
				{
					Name:     "unreal",
					Port:     7777,
					Protocol: "UDP",
				},
			},
			Type: "NodePort",
		},
	}
	service, err := serviceClient.Create(ctx, serviceResource, metaV1.CreateOptions{})
	if err != nil {
		return fmt.Errorf("failed to create service: %v", err)
	}

	if service != nil {
		// Get service port and update the game server record in the database
		for _, port := range service.Spec.Ports {
			if port.Name == "unreal" && port.Protocol == "UDP" {
				err = SetGameServerPort(ctx, id, port.NodePort)
				if err != nil {
					return fmt.Errorf("failed to update game server record: %v", err)
				}
				break
			}
		}
	}

	return nil
}

func createGameServerServiceClusterResourceWithId(ctx context.Context, id uuid.UUID) (int32, error) {
	namespace, ok := ctx.Value("namespace").(string)
	if !ok {
		return 0, fmt.Errorf("namespace not found in context")
	}

	config, ok := ctx.Value("config").(*rest.Config)
	if !ok {
		return 0, fmt.Errorf("config not found in context")
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return 0, fmt.Errorf("failed to create a clientset: %v", err)
	}

	resourceName := getResourceName(id)

	serviceClient := clientset.CoreV1().Services(namespace)
	serviceResource := &apiV1.Service{
		ObjectMeta: metaV1.ObjectMeta{
			Name: resourceName,
			Labels: map[string]string{
				"app": resourceName,
			},
		},
		Spec: apiV1.ServiceSpec{
			Selector: map[string]string{
				"app": resourceName,
			},
			Ports: []apiV1.ServicePort{
				{
					Name:     "unreal",
					Port:     7777,
					Protocol: "UDP",
				},
			},
			Type: "NodePort",
		},
	}
	s, err := serviceClient.Create(ctx, serviceResource, metaV1.CreateOptions{})
	if err != nil {
		return 0, fmt.Errorf("failed to create service: %v", err)
	}

	for _, port := range s.Spec.Ports {
		if port.Name == "unreal" && port.Protocol == "UDP" {
			return port.NodePort, nil
		}
	}

	return 0, nil
}

func deleteGameServerServiceClusterResource(ctx context.Context, id uuid.UUID) error {
	clientset := ctx.Value("clientset").(*kubernetes.Clientset)
	namespace := ctx.Value("namespace").(string)

	resourceName := id.String()

	serviceClient := clientset.CoreV1().Services(namespace)

	err := serviceClient.Delete(ctx, resourceName, metaV1.DeleteOptions{})
	if err != nil {
		return fmt.Errorf("failed to delete service: %v", err)
	}

	return nil
}
