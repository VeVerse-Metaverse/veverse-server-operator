package main

import (
	"context"
	"fmt"
	"github.com/gofrs/uuid"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/dynamic/dynamicinformer"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/cache"
	"os"
	"strconv"
	"time"
)

const (
	EnvApiV1Root           = "VE_API_ROOT_URL"
	EnvApiV2Root           = "VE_API2_ROOT_URL"
	EnvServerApiV1Key      = "VE_SERVER_API_KEY"
	EnvServerApiV2Email    = "VE_SERVER_API_EMAIL"
	EnvServerApiV2Password = "VE_SERVER_API_PASSWORD"
	EnvServerId            = "VE_SERVER_ID"
	EnvServerHost          = "VE_SERVER_HOST"
	EnvServerName          = "VE_SERVER_NAME"
	EnvServerMaxPlayers    = "VE_SERVER_MAX_PLAYERS"
	EnvServerWorldId       = "VE_SERVER_SPACE_ID"
	EnvServerAppId         = "VE_SERVER_APP_ID"
	EnvServerReleaseId     = "VE_SERVER_RELEASE_ID"
)

const (
	GameServerStatusOffline  = "offline"
	GameServerStatusOnline   = "online"
	GameServerStatusStarting = "starting"
	GameServerStatusCreated  = "created"
	GameServerStatusError    = "error"
)

func int32Ptr(i int32) *int32 { return &i }

func getResourceName(id uuid.UUID) string {
	return fmt.Sprintf("gs-%s", id.String())
}

var updateInterval = 60 * time.Second

func main() {
	// algorithm
	// 1. watch create and delete events for gameserver resources
	// 2. get all gameserver, deployment and service resources
	// 3. get all gameserver records
	// 4. check if gameserver record has a matching resources and create them as required
	// 5. update the game server metadata with the service node port

	// get update interval from env or use default value of 60 seconds
	if updateIntervalEnv := os.Getenv("UPDATE_INTERVAL"); updateIntervalEnv != "" {
		parsedInterval, err := time.ParseDuration(updateIntervalEnv)
		if err != nil {
			updateInterval = parsedInterval
		} else {
			parsedInterval, err := strconv.Atoi(updateIntervalEnv)
			if err != nil {
				updateInterval = time.Duration(parsedInterval) * time.Second
			}
		}

		if updateInterval == 0 {
			updateInterval = 60 * time.Second
		}
	}

	Logger.Infof("update interval: %v", updateInterval)

	// create a context, contains database connection, kubernetes clientset, game server resource definition, and config
	ctx := context.Background()

	//region Database

	ctx, err := DatabaseOpen(ctx)
	if err != nil {
		Logger.Fatalf("failed to setup database: %v", err)
	}

	defer func(ctx context.Context) {
		err := DatabaseClose(ctx)
		if err != nil {
			Logger.Fatalf("failed to shutdown database: %v", err)
		}
	}(ctx)

	//endregion

	//region Kubernetes Cluster Namespace
	namespace := os.Getenv("NAMESPACE")
	if namespace == "" {
		namespace = "default"
	}

	ctx = context.WithValue(ctx, "namespace", namespace)
	//endregion

	//region K8s config

	// mount config from the pod inside the cluster, required service account configuration is provided in the deployment
	config, err := rest.InClusterConfig()
	if err != nil {
		Logger.Fatalf("failed to create a config: %v", err)
	}

	// add config to context
	ctx = context.WithValue(ctx, "config", config)

	//endregion

	//region K8s client

	// create a static clientset
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		Logger.Fatalf("failed to create a clientset: %v", err)
	}

	ctx = context.WithValue(ctx, "clientset", clientset)

	// create a dynamic client using the config
	dynamicClient, err := dynamic.NewForConfig(config)
	if err != nil {
		Logger.Fatalf("failed to create a clientset: %v", err)
	}

	ctx = context.WithValue(ctx, "dynamicClient", dynamicClient)

	//endregion

	gameServerResource := schema.GroupVersionResource{Group: "veverse.com", Version: "v1", Resource: "gameservers"}

	ctx = context.WithValue(ctx, "gameServerResource", gameServerResource)

	// create a informer for the gameserver resource
	fac := dynamicinformer.NewFilteredDynamicSharedInformerFactory(dynamicClient, 0, namespace, nil)
	informer := fac.ForResource(gameServerResource).Informer()

	// add event handlers
	_, err = informer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		// handle new gameserver resources added to the cluster
		AddFunc: func(obj interface{}) {
			Logger.Infof("add event: %+v", obj)

			gameServerMetadata, ok := obj.(*unstructured.Unstructured)
			if !ok {
				Logger.Errorf("failed to convert to unstructured: %v", obj)
				return
			}

			err := createGameServerDeploymentClusterResource(ctx, *gameServerMetadata)
			if err != nil {
				Logger.Errorf("failed to create deployment: %v", err)
				return
			}

			err = createGameServerServiceClusterResource(ctx, *gameServerMetadata)
			if err != nil {
				Logger.Errorf("failed to create service: %v", err)
				return
			}
		},
		UpdateFunc: func(oldObj, newObj interface{}) {
			Logger.Infof("update event: %+v", newObj)
		},
		// handle gameserver resources removed from the cluster
		DeleteFunc: func(obj interface{}) {
			Logger.Infof("delete event: %+v", obj)

			gameServerMetadata, ok := obj.(*unstructured.Unstructured)
			if !ok {
				Logger.Errorf("failed to convert to unstructured: %v", obj)
				return
			}

			id, err := uuid.FromString(gameServerMetadata.GetName())
			if err != nil {
				Logger.Errorf("failed to parse id: %v", err)
				return
			}

			err = deleteGameServerDeploymentClusterResource(ctx, id)
			if err != nil {
				Logger.Errorf("failed to delete deployment: %v", err)
				return
			}

			err = deleteGameServerServiceClusterResource(ctx, id)
			if err != nil {
				Logger.Errorf("failed to delete service: %v", err)
				return
			}
		},
	})
	if err != nil {
		Logger.Fatalf("failed to add event handler: %v", err)
	}

	// get all current game server resources and create deployments and services for them if they don't exist
	for {
		gameServerRecords, err := GetOnlineGameServers(ctx)
		if err != nil {
			Logger.Errorf("failed to get active game servers: %v", err)
			continue
		}

		// check for matching deployments and services which need to be created or deleted if the game server resource is offline or in error state and were not handled by the event handler for some reason
		for _, gameServerRecord := range gameServerRecords.Entities {
			if gameServerRecord.Status == GameServerStatusOffline || gameServerRecord.Status == GameServerStatusError {
				// delete offline and error game servers
				gameServer, err := getGameServerClusterResource(ctx, gameServerRecord.Id)
				if err != nil {
					Logger.Errorf("failed to get game server: %v", err)
					continue
				}
				if gameServer == nil {
					err := deleteGameServerClusterResource(ctx, gameServerRecord.Id)
					if err != nil {
						Logger.Errorf("failed to delete game server: %v", err)
						continue
					}
				}

				// delete deployment if it still exists
				deployment, err := getGameServerDeploymentClusterResource(ctx, gameServerRecord.Id)
				if err != nil {
					Logger.Errorf("failed to get deployment: %v", err)
					continue
				}
				if deployment != nil {
					err := deleteGameServerDeploymentClusterResource(ctx, gameServerRecord.Id)
					if err != nil {
						Logger.Errorf("failed to delete deployment: %v", err)
						continue
					}
				}

				// delete service if it still exists
				service, err := getGameServerServiceClusterResource(ctx, gameServerRecord.Id)
				if err != nil {
					Logger.Errorf("failed to get service: %v", err)
					continue
				}
				if service != nil {
					err := deleteGameServerServiceClusterResource(ctx, gameServerRecord.Id)
					if err != nil {
						Logger.Errorf("failed to delete service: %v", err)
						continue
					}
				}
			} else if gameServerRecord.Status == GameServerStatusOnline || gameServerRecord.Status == GameServerStatusStarting || gameServerRecord.Status == GameServerStatusCreated {
				// if there is no deployment, but we have an active game server record, mark the game server as offline and delete the matching service to release node ports
				deployment, err := getGameServerDeploymentClusterResource(ctx, gameServerRecord.Id)
				if err != nil {
					Logger.Errorf("failed to get deployment: %v", err)
					continue
				}
				if deployment == nil {
					Logger.Warningf("deployment not found for game server: %v, marking server as offline", gameServerRecord.Id)

					err = SetGameServerOffline(ctx, gameServerRecord.Id)
					if err != nil {
						Logger.Errorf("failed to set game server offline: %v", err)
						continue
					}

					// check if there is a service for the game server and delete it
					service, err := getGameServerServiceClusterResource(ctx, gameServerRecord.Id)
					if err != nil {
						Logger.Errorf("failed to get service: %v", err)
						continue
					}

					if service != nil {
						err = deleteGameServerServiceClusterResource(ctx, gameServerRecord.Id)
						if err != nil {
							Logger.Errorf("failed to delete service: %v", err)
							continue
						}
					}

					continue
				}

				// if there is no service, but server exists, try to create a service for it
				service, err := getGameServerServiceClusterResource(ctx, gameServerRecord.Id)
				if err != nil {
					Logger.Errorf("failed to get service: %v", err)
					continue
				}
				if service == nil {
					port, err := createGameServerServiceClusterResourceWithId(ctx, gameServerRecord.Id)
					if err != nil {
						Logger.Errorf("failed to create service: %v", err)
						continue
					}

					// update the game server record with the port
					err = SetGameServerPort(ctx, gameServerRecord.Id, port)
					if err != nil {
						Logger.Errorf("failed to set game server port: %v", err)
						continue
					}
				}
			}
		}

		time.Sleep(updateInterval)
	}
}
