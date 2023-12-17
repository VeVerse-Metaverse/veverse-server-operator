package main

import (
	"context"
	"github.com/gofrs/uuid"
	metaV1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
)

func getGameServerClusterResource(ctx context.Context, id uuid.UUID) (*unstructured.Unstructured, error) {
	namespace := ctx.Value("namespace").(string)
	dynamicClient := ctx.Value("dynamicClient").(*dynamic.DynamicClient)
	gameServerResource := ctx.Value("gameServerResource").(schema.GroupVersionResource)

	resourceName := getResourceName(id)

	gameServer, err := dynamicClient.Resource(gameServerResource).Namespace(namespace).Get(ctx, resourceName, metaV1.GetOptions{})
	if err != nil {
		return nil, err
	}

	return gameServer, nil
}

func deleteGameServerClusterResource(ctx context.Context, id uuid.UUID) error {
	namespace := ctx.Value("namespace").(string)
	dynamicClient := ctx.Value("dynamicClient").(*dynamic.DynamicClient)
	gameServerResource := ctx.Value("gameServerResource").(schema.GroupVersionResource)

	resourceName := getResourceName(id)

	err := dynamicClient.Resource(gameServerResource).Namespace(namespace).Delete(ctx, resourceName, metaV1.DeleteOptions{})
	if err != nil {
		return err
	}

	return nil
}
