package main

import (
	"context"
	"fmt"
	"github.com/gofrs/uuid"
	appsV1 "k8s.io/api/apps/v1"
	apiV1 "k8s.io/api/core/v1"
	metaV1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/client-go/kubernetes"
)

func getGameServerDeploymentClusterResource(ctx context.Context, id uuid.UUID) (*appsV1.Deployment, error) {
	clientset := ctx.Value("clientset").(*kubernetes.Clientset)
	namespace := ctx.Value("namespace").(string)

	resourceName := getResourceName(id)

	deploymentsClient := clientset.AppsV1().Deployments(namespace)

	deployment, err := deploymentsClient.Get(ctx, resourceName, metaV1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to get deployment: %v", err)
	}

	return deployment, nil
}

func createGameServerDeploymentClusterResource(ctx context.Context, metadata unstructured.Unstructured) error {
	namespace, ok := ctx.Value("namespace").(string)
	if !ok {
		return fmt.Errorf("namespace not found in context")
	}

	clientset, ok := ctx.Value("clientset").(*kubernetes.Clientset)
	if !ok {
		return fmt.Errorf("clientset not found in context")
	}

	Logger.Infof("metadata: %v", metadata)

	//region Specification
	spec, ok := metadata.Object["spec"].(map[string]interface{})
	if !ok {
		return fmt.Errorf("spec not found in game server metadata")
	}

	//region ID
	if _, ok := spec["id"].(string); !ok {
		return fmt.Errorf("id not found in game server metadata")
	}

	id, err := uuid.FromString(spec["id"].(string))
	if err != nil {
		return fmt.Errorf("failed to parse id: %v", err)
	}

	resourceName := getResourceName(id)
	//endregion

	//region Environment Variables
	var envs []apiV1.EnvVar
	envSpec, ok := spec["env"].([]interface{})
	if !ok {
		return fmt.Errorf("envs not found in game server metadata")
	}

	for _, envMetadata := range envSpec {
		env := envMetadata.(map[string]interface{})
		envs = append(envs, apiV1.EnvVar{
			Name:  env["name"].(string),
			Value: env["value"].(string),
		})
	}

	envs = append(envs, apiV1.EnvVar{Name: EnvServerId, Value: resourceName})
	envs = append(envs, apiV1.EnvVar{Name: EnvServerName, Value: resourceName})
	//endregion

	//region Settings
	settingSpec, ok := spec["settings"].(map[string]interface{})
	if !ok {
		return fmt.Errorf("settings not found in game server metadata")
	}

	//region API
	apiSettingSpec, ok := settingSpec["api"].(map[string]interface{})
	if !ok {
		return fmt.Errorf("api settings not found in game server metadata")
	}

	//region V1
	apiV1SettingSpec, ok := apiSettingSpec["v1"].(map[string]interface{})
	if !ok {
		return fmt.Errorf("v1 api settings not found in game server metadata")
	}
	apiV1Root, ok := apiV1SettingSpec["url"].(string)
	if !ok {
		return fmt.Errorf("v1 api root not found in game server metadata")
	}

	envs = append(envs, apiV1.EnvVar{Name: EnvApiV1Root, Value: apiV1Root})

	apiV1Key, ok := apiV1SettingSpec["key"].(string)
	if !ok {
		return fmt.Errorf("v1 api key not found in game server metadata")
	}

	envs = append(envs, apiV1.EnvVar{Name: EnvServerApiV1Key, Value: apiV1Key})
	//endregion

	//region V2

	apiV2SettingSpec, ok := apiSettingSpec["v2"].(map[string]interface{})
	if !ok {
		return fmt.Errorf("v2 api settings not found in game server metadata")
	}

	apiV2Root, ok := apiV2SettingSpec["url"].(string)
	if !ok {
		return fmt.Errorf("v2 api root not found in game server metadata")
	}

	envs = append(envs, apiV1.EnvVar{Name: EnvApiV2Root, Value: apiV2Root})

	apiV2Email, ok := apiV2SettingSpec["email"].(string)
	if !ok {
		return fmt.Errorf("v2 api email not found in game server metadata")
	}

	envs = append(envs, apiV1.EnvVar{Name: EnvServerApiV2Email, Value: apiV2Email})

	apiV2Password, ok := apiV2SettingSpec["password"].(string)
	if !ok {
		return fmt.Errorf("v2 api password not found in game server metadata")
	}

	envs = append(envs, apiV1.EnvVar{Name: EnvServerApiV2Password, Value: apiV2Password})

	//endregion

	//endregion

	//region App
	appSpec, ok := settingSpec["app"].(map[string]interface{})
	if !ok {
		return fmt.Errorf("app not found in game server metadata")
	}

	appId, ok := appSpec["id"].(string)
	if !ok {
		return fmt.Errorf("app id not found in game server metadata")
	}

	envs = append(envs, apiV1.EnvVar{Name: EnvServerAppId, Value: appId})
	//endregion

	//region Release
	releaseSpec, ok := settingSpec["release"].(map[string]interface{})
	if !ok {
		return fmt.Errorf("release not found in game server metadata")
	}

	releaseId, ok := releaseSpec["id"].(string)
	if !ok {
		return fmt.Errorf("release id not found in game server metadata")
	}

	envs = append(envs, apiV1.EnvVar{Name: EnvServerReleaseId, Value: releaseId})
	//endregion

	//region Players
	playerSettingSpec, ok := settingSpec["players"].(map[string]interface{})
	if !ok {
		return fmt.Errorf("player settings not found in game server metadata")
	}
	Logger.Infof("type of playerSettingSpec: %T", playerSettingSpec["max"])
	maxPlayers, ok := playerSettingSpec["max"].(int64)
	if !ok {
		return fmt.Errorf("max players not found in game server metadata")
	}
	envs = append(envs, apiV1.EnvVar{Name: EnvServerMaxPlayers, Value: fmt.Sprintf("%d", maxPlayers)})
	//endregion

	//region World
	worldSettingSpec, ok := settingSpec["world"].(map[string]interface{})
	if !ok {
		return fmt.Errorf("world settings not found in game server metadata")
	}
	worldId, ok := worldSettingSpec["id"].(string)
	if !ok {
		return fmt.Errorf("world id not found in game server metadata")
	}
	envs = append(envs, apiV1.EnvVar{Name: EnvServerWorldId, Value: worldId})
	//endregion

	//region Server
	serverSettingSpec, ok := settingSpec["server"].(map[string]interface{})
	if !ok {
		return fmt.Errorf("server settings not found in game server metadata")
	}

	//region Host
	serverHost, ok := serverSettingSpec["host"].(string)
	if !ok {
		return fmt.Errorf("server host not found in game server metadata")
	}
	envs = append(envs, apiV1.EnvVar{Name: EnvServerHost, Value: serverHost})
	//endregion

	//region Container Image
	serverImage, ok := serverSettingSpec["image"].(string)
	if !ok {
		return fmt.Errorf("server image not found in game server metadata")
	}

	serverImagePullSecretsSpec, ok := serverSettingSpec["imagePullSecrets"].([]interface{})
	if !ok {
		return fmt.Errorf("server image pull secrets not found in game server metadata")
	}
	serverImagePullSecrets := make([]apiV1.LocalObjectReference, len(serverImagePullSecretsSpec))
	for i, serverImagePullSecret := range serverImagePullSecretsSpec {
		name, ok := serverImagePullSecret.(string)
		if !ok {
			return fmt.Errorf("server image pull secret is not a string")
		}
		serverImagePullSecrets[i] = apiV1.LocalObjectReference{Name: name}
	}
	//endregion

	//endregion

	//endregion

	//endregion

	deploymentsClient := clientset.AppsV1().Deployments(namespace)
	deploymentResource := &appsV1.Deployment{
		ObjectMeta: metaV1.ObjectMeta{
			Name: resourceName,
			Labels: map[string]string{
				"app": resourceName,
			},
		},
		Spec: appsV1.DeploymentSpec{
			Replicas: int32Ptr(1),
			Selector: &metaV1.LabelSelector{
				MatchLabels: map[string]string{
					"app": resourceName,
				},
			},
			Template: apiV1.PodTemplateSpec{
				ObjectMeta: metaV1.ObjectMeta{
					Labels: map[string]string{
						"app": resourceName,
					},
				},
				Spec: apiV1.PodSpec{
					ImagePullSecrets: serverImagePullSecrets,
					Containers: []apiV1.Container{
						{
							Name:  resourceName,
							Env:   envs,
							Image: serverImage,
							Ports: []apiV1.ContainerPort{
								{
									Name:          "unreal",
									ContainerPort: 7777,
									Protocol:      "UDP",
								},
							},
						},
					},
				},
			},
		},
	}
	_, err = deploymentsClient.Create(ctx, deploymentResource, metaV1.CreateOptions{})
	if err != nil {
		return fmt.Errorf("failed to create deployment: %v", err)
	}

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
	_, err = serviceClient.Create(ctx, serviceResource, metaV1.CreateOptions{})
	if err != nil {
		return fmt.Errorf("failed to create service: %v", err)
	}

	return nil
}

func deleteGameServerDeploymentClusterResource(ctx context.Context, id uuid.UUID) error {
	clientset := ctx.Value("clientset").(*kubernetes.Clientset)
	namespace := ctx.Value("namespace").(string)

	resourceName := getResourceName(id)

	deploymentsClient := clientset.AppsV1().Deployments(namespace)

	err := deploymentsClient.Delete(ctx, resourceName, metaV1.DeleteOptions{})
	if err != nil {
		return fmt.Errorf("failed to delete deployment: %v", err)
	}

	return nil
}
