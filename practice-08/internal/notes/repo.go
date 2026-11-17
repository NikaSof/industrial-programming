package notes

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

var (
	ErrNotFound = errors.New("note not found")
	ErrConflict = errors.New("note title already exists")
	ErrInvalid  = errors.New("invalid note data")
)

type Repo struct {
	col *mongo.Collection
}

func NewRepo(ctx context.Context, database *mongo.Database) (*Repo, error) {
	col := database.Collection("notes")
	indexCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	_, err := col.Indexes().CreateOne(indexCtx, mongo.IndexModel{
		Keys:    bson.D{{Key: "title", Value: 1}},
		Options: options.Index().SetUnique(true),
	})
	if err != nil {
		return nil, fmt.Errorf("create notes index: %w", err)
	}
	return &Repo{col: col}, nil
}

func (r *Repo) Create(ctx context.Context, title, content string) (Note, error) {
	title = strings.TrimSpace(title)
	if title == "" {
		return Note{}, ErrInvalid
	}
	now := time.Now().UTC().Truncate(time.Millisecond)
	note := Note{
		ID: bson.NewObjectID(), Title: title, Content: content,
		CreatedAt: now, UpdatedAt: now,
	}
	if _, err := r.col.InsertOne(ctx, note); err != nil {
		return Note{}, repoError("create note", err)
	}
	return note, nil
}

func (r *Repo) ByID(ctx context.Context, idHex string) (Note, error) {
	id, err := bson.ObjectIDFromHex(idHex)
	if err != nil {
		return Note{}, ErrNotFound
	}
	var note Note
	if err := r.col.FindOne(ctx, bson.M{"_id": id}).Decode(&note); err != nil {
		return Note{}, repoError("find note", err)
	}
	return note, nil
}

func (r *Repo) List(ctx context.Context, q string, limit, skip int64) ([]Note, error) {
	if limit <= 0 || limit > 200 {
		limit = 20
	}
	if skip < 0 {
		skip = 0
	}
	filter := bson.M{}
	if q != "" {
		// Пользователь вводит текст поиска, а не произвольное регулярное выражение.
		filter["title"] = bson.M{"$regex": regexp.QuoteMeta(q), "$options": "i"}
	}
	opts := options.Find().SetLimit(limit).SetSkip(skip).SetSort(bson.D{
		{Key: "createdAt", Value: -1},
		{Key: "_id", Value: -1}, // Однозначный порядок при одинаковом времени создания.
	})
	cur, err := r.col.Find(ctx, filter, opts)
	if err != nil {
		return nil, repoError("list notes", err)
	}
	defer func() {
		cleanupCtx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()
		_ = cur.Close(cleanupCtx)
	}()

	items := make([]Note, 0)
	for cur.Next(ctx) {
		var note Note
		if err := cur.Decode(&note); err != nil {
			return nil, repoError("decode note", err)
		}
		items = append(items, note)
	}
	if err := cur.Err(); err != nil {
		return nil, repoError("read notes cursor", err)
	}
	return items, nil
}

func (r *Repo) Update(ctx context.Context, idHex string, title, content *string) (Note, error) {
	id, err := bson.ObjectIDFromHex(idHex)
	if err != nil {
		return Note{}, ErrNotFound
	}
	if title == nil && content == nil {
		return Note{}, ErrInvalid
	}
	set := bson.M{"updatedAt": time.Now().UTC().Truncate(time.Millisecond)}
	if title != nil {
		trimmed := strings.TrimSpace(*title)
		if trimmed == "" {
			return Note{}, ErrInvalid
		}
		set["title"] = trimmed
	}
	if content != nil {
		set["content"] = *content
	}
	var note Note
	err = r.col.FindOneAndUpdate(ctx, bson.M{"_id": id}, bson.M{"$set": set},
		options.FindOneAndUpdate().SetReturnDocument(options.After)).Decode(&note)
	if err != nil {
		return Note{}, repoError("update note", err)
	}
	return note, nil
}

func (r *Repo) Delete(ctx context.Context, idHex string) error {
	id, err := bson.ObjectIDFromHex(idHex)
	if err != nil {
		return ErrNotFound
	}
	result, err := r.col.DeleteOne(ctx, bson.M{"_id": id})
	if err != nil {
		return repoError("delete note", err)
	}
	if result.DeletedCount == 0 {
		return ErrNotFound
	}
	return nil
}

func repoError(operation string, err error) error {
	if errors.Is(err, mongo.ErrNoDocuments) {
		return ErrNotFound
	}
	if mongo.IsDuplicateKeyError(err) {
		return ErrConflict
	}
	return fmt.Errorf("%s: %w", operation, err)
}
