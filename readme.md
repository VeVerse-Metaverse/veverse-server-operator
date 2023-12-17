* Operator runs as a pod in corresponding game server namespace.
* It is responsible for watching custom resources `gameservers.veverse.com` and manage deployments and services based on
  those resources.
* It must be deployed to the namespace where gameserver resources for corresponding environment are created.
* Game server resources created and deleted by the API on client request or other conditions. Each game server resource
  gets an unique ID (UUIDv4).
* When the gameserver is created, the operator will create a deployment with a single pod and a service to provide the
  UDP port for the deployment's pod. It will also update the game server within the database and mark it as "starting".
* When the pod is started and is ready, it will update the game server record within the database and mark it as online.
  Then the game server binary within the pod will send constant updates to the database to keep the game server
  online.
* When the pod is stopped gracefully, it will update the game server within the database and mark it as "offline". If
  the
  pod is stopped unexpectedly, it would not update it state, so clients should check the updated at time to see if the
  game server is still online.
* When the gameserver resource is deleted, the operator will delete the deployment and service for the game server.
* The operator monitors deployments and services and deletes them if they are not matching any game server
  resource.