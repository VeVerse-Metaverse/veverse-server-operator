package main

import (
	"context"
	vModel "dev.hackerman.me/artheon/veverse-shared/model"
	"fmt"
	"github.com/gofrs/uuid"
	"github.com/jackc/pgtype"
	pgtypeuuid "github.com/jackc/pgtype/ext/gofrs-uuid"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/log/logrusadapter"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/sirupsen/logrus"
	"os"
)

func DatabaseOpen(ctx context.Context) (context.Context, error) {
	host := os.Getenv("DATABASE_HOST")
	port := os.Getenv("DATABASE_PORT")
	user := os.Getenv("DATABASE_USER")
	pass := os.Getenv("DATABASE_PASS")
	name := os.Getenv("DATABASE_NAME")

	url := fmt.Sprintf("postgres://%s:%s@%s:%s/%s", user, pass, host, port, name)

	config, err := pgxpool.ParseConfig(url)
	if err != nil {
		Logger.Fatal(err)
	}

	config.AfterConnect = func(ctx context.Context, conn *pgx.Conn) error {
		conn.ConnInfo().RegisterDataType(pgtype.DataType{
			Value: &pgtypeuuid.UUID{},
			Name:  "uuid",
			OID:   pgtype.UUIDOID,
		})
		return nil
	}

	logger := &logrus.Logger{
		Out:          os.Stderr,
		Formatter:    new(logrus.JSONFormatter),
		Hooks:        make(logrus.LevelHooks),
		Level:        logrus.InfoLevel,
		ExitFunc:     os.Exit,
		ReportCaller: false,
	}

	env := os.Getenv("ENVIRONMENT")
	if env != "prod" {
		config.ConnConfig.Logger = logrusadapter.NewLogger(logger)
	}

	pool, err := pgxpool.ConnectConfig(ctx, config)
	if err != nil {
		return ctx, fmt.Errorf("unable to connect to database: %v", err)
	}

	ctx = context.WithValue(ctx, "database", pool)

	return ctx, nil
}

func DatabaseClose(ctx context.Context) error {
	db, ok := ctx.Value("database").(*pgxpool.Pool)
	if !ok {
		return fmt.Errorf("unable to get database connection")
	}

	db.Close()

	return nil
}

func GetApps(ctx context.Context) ([]vModel.AppV2, error) {
	var apps []vModel.AppV2

	db, ok := ctx.Value("database").(*pgxpool.Pool)
	if !ok {
		return apps, fmt.Errorf("unable to get database connection")
	}

	rows, err := db.Query(ctx, "SELECT id, name, external FROM apps")
	if err != nil {
		return apps, err
	}

	for rows.Next() {
		var app vModel.AppV2
		err := rows.Scan(&app.Id, &app.Name, &app.External)
		if err != nil {
			return apps, err
		}

		apps = append(apps, app)
	}

	return apps, nil
}

func GetReleases(ctx context.Context) ([]vModel.ReleaseV2, error) {
	var releases []vModel.ReleaseV2

	db, ok := ctx.Value("database").(*pgxpool.Pool)
	if !ok {
		return releases, fmt.Errorf("unable to get database connection")
	}

	rows, err := db.Query(ctx, "select e.id, e.created_at, e.updated_at, r.entity_id, r.name, r.version from release_v2 r left join entities e on r.id = e.id")
	if err != nil {
		return releases, err
	}

	for rows.Next() {
		var release vModel.ReleaseV2
		err := rows.Scan(
			&release.Id,
			&release.CreatedAt,
			&release.UpdatedAt,
			&release.EntityId,
			&release.Name,
			&release.Version)
		if err != nil {
			return releases, err
		}

		releases = append(releases, release)
	}

	return releases, nil
}

func GetOnlineGameServers(ctx context.Context) (vModel.GameServerV2Batch, error) {
	var servers vModel.GameServerV2Batch

	db, ok := ctx.Value("database").(*pgxpool.Pool)
	if !ok {
		return servers, fmt.Errorf("unable to get database connection")
	}

	// query for the total number of online game servers that have been updated in the last minute (did not time out)
	row := db.QueryRow(ctx, `select count(*) from game_server_v2 s left join entities e on s.id = e.id where s.status = 'online' and e.updated_at >= now() - interval '1 minute'`)
	err := row.Scan(&servers.Total)
	if err != nil {
		return servers, err
	}

	// query for all online game servers that have been updated in the last minute (did not time out)
	rows, err := db.Query(ctx, `select e.id, 
       e.created_at, 
       e.updated_at, 
       e.public, 
       s.release_id,
       s.world_id,
       s.game_mode_id,
       s.region_id,
       s.type,
       s.host,
       s.port,
       s.max_players,
       s.status,
       s.status_message
from game_server_v2 s
left join entities e on s.id = e.id
where status = 'online' and e.updated_at >= now() - interval '1 minute'`)
	if err != nil {
		return servers, err
	}

	for rows.Next() {
		var server vModel.GameServerV2
		err := rows.Scan(
			&server.Id,
			&server.CreatedAt,
			&server.UpdatedAt,
			&server.Public,
			&server.ReleaseId,
			&server.WorldId,
			&server.GameModeId,
			&server.RegionId,
			&server.Type,
			&server.Host,
			&server.Port,
			&server.MaxPlayers,
			&server.Status,
			&server.StatusMessage)
		if err != nil {
			return servers, err
		}

		servers.Entities = append(servers.Entities, server)
	}

	return servers, nil
}

func SetGameServerOffline(ctx context.Context, id uuid.UUID) error {
	db, ok := ctx.Value("database").(*pgxpool.Pool)
	if !ok {
		return fmt.Errorf("unable to get database connection")
	}

	_, err := db.Exec(ctx, `update game_server_v2 set status = 'offline' where id = $1`, id)
	if err != nil {
		return fmt.Errorf("unable to set server offline: %v", err)
	}

	_, err = db.Exec(ctx, `update entities set updated_at = now() where id = $1`, id)
	if err != nil {
		return fmt.Errorf("unable to update entity updated_at: %v", err)
	}

	return nil
}

func SetGameServerPort(ctx context.Context, id uuid.UUID, port int32) error {
	db, ok := ctx.Value("database").(*pgxpool.Pool)
	if !ok {
		return fmt.Errorf("unable to get database connection")
	}

	_, err := db.Exec(ctx, `update game_server_v2 set port = $1 where id = $2`, port, id)
	if err != nil {
		return fmt.Errorf("unable to set server port: %v", err)
	}

	_, err = db.Exec(ctx, `update entities set updated_at = now() where id = $1`, id)
	if err != nil {
		return fmt.Errorf("unable to update entity updated_at: %v", err)
	}

	return nil
}
