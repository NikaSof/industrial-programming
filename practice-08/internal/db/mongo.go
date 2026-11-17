package db

import (
	"context"
	"errors"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
	"go.mongodb.org/mongo-driver/v2/mongo/readpref"
)

type MongoDeps struct {
	Client   *mongo.Client
	Database *mongo.Database
}

func ConnectMongo(ctx context.Context, uri, dbName string) (*MongoDeps, error) {
	if uri == "" || dbName == "" {
		return nil, errors.New("MongoDB URI and database name are required")
	}

	opts := options.Client().ApplyURI(uri).
		SetConnectTimeout(10 * time.Second).
		SetServerSelectionTimeout(5 * time.Second)
	client, err := mongo.Connect(opts)
	if err != nil {
		return nil, fmt.Errorf("create MongoDB client: %w", err)
	}

	// Connect создаёт клиент, а Ping проверяет доступность сервера и аутентификацию.
	pingCtx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()
	if err := client.Ping(pingCtx, readpref.Primary()); err != nil {
		// Очистка не должна зависеть от уже отменённого контекста подключения.
		cleanupCtx, cleanupCancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cleanupCancel()
		if closeErr := client.Disconnect(cleanupCtx); closeErr != nil {
			return nil, errors.Join(
				fmt.Errorf("ping MongoDB: %w", err),
				fmt.Errorf("close MongoDB client: %w", closeErr),
			)
		}
		return nil, fmt.Errorf("ping MongoDB: %w", err)
	}

	return &MongoDeps{
		Client:   client,
		Database: client.Database(dbName),
	}, nil
}
