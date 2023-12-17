package main

//
//type Identifier struct {
//	Id *uuid.UUID `json:"id,omitempty"`
//}
//
//type Timestamps struct {
//	CreatedAt *time.Time `json:"createdAt,omitempty"`
//	UpdatedAt *time.Time `json:"updatedAt,omitempty"`
//}
//
//type Entity struct {
//	Identifier
//
//	EntityType *string `json:"entityType,omitempty"`
//	Public     *bool   `json:"public,omitempty"`
//	Views      *int32  `json:"views,omitempty"`
//
//	Timestamps
//
//	Owner       *User        `json:"owner,omitempty"`
//	Accessibles []Accessible `json:"accessibles,omitempty"`
//	Files       []File       `json:"files,omitempty"`
//}
//
//type EntityTrait struct {
//	Identifier
//	EntityId *uuid.UUID `json:"entityId,omitempty"`
//}
//
//type Accessible struct {
//	EntityTrait
//
//	UserId    uuid.UUID `json:"userId"`
//	IsOwner   bool      `json:"isOwner"`
//	CanView   bool      `json:"canView"`
//	CanEdit   bool      `json:"canEdit"`
//	CanDelete bool      `json:"canDelete"`
//
//	Timestamps
//}
//
//type File struct {
//	EntityTrait
//
//	Type         string     `json:"type"`
//	Url          string     `json:"url"`
//	Mime         *string    `json:"mime,omitempty"`
//	Size         *int64     `json:"size,omitempty"`
//	Version      int        `json:"version,omitempty"`        // version of the file if versioned
//	Deployment   string     `json:"deploymentType,omitempty"` // server or client if applicable
//	Platform     string     `json:"platform,omitempty"`       // platform if applicable
//	UploadedBy   *uuid.UUID `json:"uploadedBy,omitempty"`     // user that uploaded the file
//	Width        *int       `json:"width,omitempty"`
//	Height       *int       `json:"height,omitempty"`
//	CreatedAt    time.Time  `json:"createdAt,omitempty"`
//	UpdatedAt    *time.Time `json:"updatedAt,omitempty"`
//	Index        int        `json:"variation,omitempty"`    // variant of the file if applicable (e.g. PDF pages)
//	OriginalPath *string    `json:"originalPath,omitempty"` // original relative path to maintain directory structure (e.g. for releases)
//
//	Timestamps
//}
//
//type User struct {
//	Entity
//
//	Email       *string    `json:"email,omitempty"`
//	Name        *string    `json:"name"`
//	IsActive    bool       `json:"isActive,omitempty"`
//	IsAdmin     bool       `json:"isAdmin,omitempty"`
//	IsMuted     bool       `json:"isMuted,omitempty"`
//	IsBanned    bool       `json:"isBanned,omitempty"`
//	IsInternal  bool       `json:"isInternal,omitempty"`
//	ActivatedAt *time.Time `json:"activatedAt,omitempty"`
//	AllowEmails bool       `json:"allowEmails,omitempty"`
//}
//
//// Release struct
//type Release struct {
//	Entity
//
//	AppId          *uuid.UUID `json:"appId,omitempty"`
//	AppName        string     `json:"appName,omitempty"`
//	AppTitle       string     `json:"appTitle,omitempty"`
//	AppDescription *string    `json:"appDescription"`
//	AppUrl         *string    `json:"appUrl"`
//	AppExternal    *bool      `json:"appExternal"`
//	Version        string     `json:"version,omitempty"`
//	CodeVersion    string     `json:"codeVersion,omitempty"`
//	ContentVersion string     `json:"contentVersion,omitempty"`
//	Name           *string    `json:"name,omitempty"`
//	Description    *string    `json:"description,omitempty"`
//	Archive        *bool      `json:"archive"`
//}
//
//type Server struct {
//	Identifier
//	Timestamps
//	AppId          *uuid.UUID `json:"appId,omitempty"`      // app id if applicable
//	ReleaseId      *uuid.UUID `json:"releaseId,omitempty"`  // release id if applicable
//	Public         *bool      `json:"public,omitempty"`     // public or private server (default public)
//	Map            *string    `json:"map,omitempty"`        // map name running at the server (e.g. "/Template_Museum/Maps/MuseumTemplate")
//	Host           *string    `json:"host,omitempty"`       // host name or ip address (e.g. "gameserver.veverse.com")
//	Port           *int       `json:"port,omitempty"`       // port number (e.g. 7777, in cluster services usually get ports >30000 allocated)
//	WorldId        *uuid.UUID `json:"worldId,omitempty"`    // world id that runs at the server (e.g. "00000000-0000-0000-0000-000000000000")
//	MaxPlayers     *int       `json:"maxPlayers,omitempty"` // max players allowed at the server (e.g. 100)
//	GameMode       *string    `json:"gameMode,omitempty"`   // game mode class (e.g. "/VeGame/Shared/Blueprints/VeGameMode_BP.VeGameMode_BP_C")
//	UserId         *uuid.UUID `json:"userId,omitempty"`     // user id that started the server (e.g. "00000000-0000-0000-0000-000000000000")
//	Status         *string    `json:"status,omitempty"`     // server status (created, starting, online, offline)
//	ContainerImage *string    `json:"image,omitempty"`      // server container image (e.g. "registry.veverse.com/veverse/veverse-server:latest")
//}
//
//type App struct {
//	Entity
//
//	Name              string `json:"name,omitempty"`
//	Description       string `json:"description,omitempty"`
//	Url               string `json:"url,omitempty"`
//	PixelStreamingUrl string `json:"pixelStreamingUrl,omitempty"`
//	PrivacyPolicyURL  string `json:"privacyPolicyURL,omitempty"`
//	External          bool   `json:"external,omitempty"`
//	Title             string `json:"title,omitempty"` // Display name
//}
