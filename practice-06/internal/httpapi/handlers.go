package httpapi

import (
	"borovskaya/pr6/internal/models"
	"encoding/json"
	"errors"
	"github.com/go-chi/chi/v5"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"io"
	"net/http"
	"strconv"
	"strings"
	"unicode/utf8"
)

type Handlers struct{ db *gorm.DB }

func (h *Handlers) ListUsers(w http.ResponseWriter, r *http.Request) {
	users := make([]models.User, 0)
	if err := h.db.WithContext(r.Context()).Order("id ASC").Find(&users).Error; err != nil {
		failure(w, err)
		return
	}
	writeJSON(w, http.StatusOK, users)
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	b, e := json.Marshal(v)
	if e != nil {
		status = 500
		b = []byte(`{"error":"internal server error"}`)
	}
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_, _ = w.Write(b)
}
func writeErr(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, map[string]string{"error": msg})
}
func decode(w http.ResponseWriter, r *http.Request, v any) bool {
	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)
	d := json.NewDecoder(r.Body)
	d.DisallowUnknownFields()
	if d.Decode(v) != nil {
		writeErr(w, 400, "invalid JSON")
		return false
	}
	if d.Decode(new(any)) != io.EOF {
		writeErr(w, 400, "expected one JSON object")
		return false
	}
	return true
}
func failure(w http.ResponseWriter, e error) {
	switch {
	case errors.Is(e, gorm.ErrRecordNotFound):
		writeErr(w, 404, "record not found")
	case errors.Is(e, gorm.ErrDuplicatedKey):
		writeErr(w, 409, "record already exists")
	case errors.Is(e, gorm.ErrForeignKeyViolated):
		writeErr(w, 400, "invalid relation")
	default:
		writeErr(w, 500, "database operation failed")
	}
}
func (h *Handlers) Health(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, 200, map[string]string{"status": "ok"})
}
func (h *Handlers) CreateUser(w http.ResponseWriter, r *http.Request) {
	var in struct {
		Name  string `json:"name"`
		Email string `json:"email"`
	}
	if !decode(w, r, &in) {
		return
	}
	in.Name = strings.TrimSpace(in.Name)
	in.Email = strings.TrimSpace(in.Email)
	if in.Name == "" || in.Email == "" || utf8.RuneCountInString(in.Name) > 100 || utf8.RuneCountInString(in.Email) > 200 {
		writeErr(w, 400, "valid name and email are required")
		return
	}
	u := models.User{Name: in.Name, Email: in.Email}
	if e := h.db.WithContext(r.Context()).Create(&u).Error; e != nil {
		failure(w, e)
		return
	}
	writeJSON(w, 201, u)
}
func (h *Handlers) CreateNote(w http.ResponseWriter, r *http.Request) {
	var in struct {
		Title   string   `json:"title"`
		Content string   `json:"content"`
		UserID  uint     `json:"userId"`
		Tags    []string `json:"tags"`
	}
	if !decode(w, r, &in) {
		return
	}
	in.Title = strings.TrimSpace(in.Title)
	if in.Title == "" || in.UserID == 0 || utf8.RuneCountInString(in.Title) > 200 {
		writeErr(w, 400, "title and userId are required")
		return
	}
	names := []string{}
	seen := map[string]bool{}
	for _, n := range in.Tags {
		n = strings.TrimSpace(n)
		if utf8.RuneCountInString(n) > 50 {
			writeErr(w, 400, "tag too long")
			return
		}
		if n != "" && !seen[n] {
			seen[n] = true
			names = append(names, n)
		}
	}
	note := models.Note{Title: in.Title, Content: in.Content, UserID: in.UserID, Tags: []models.Tag{}}
	e := h.db.WithContext(r.Context()).Transaction(func(tx *gorm.DB) error {
		var user models.User
		if e := tx.First(&user, in.UserID).Error; e != nil {
			return e
		}
		for _, name := range names {
			tag := models.Tag{Name: name}
			// Unique index and ON CONFLICT also handle simultaneous creation of a tag.
			if e := tx.Clauses(clause.OnConflict{Columns: []clause.Column{{Name: "name"}}, DoNothing: true}).Create(&tag).Error; e != nil {
				return e
			}
			if e := tx.Where("name = ?", name).First(&tag).Error; e != nil {
				return e
			}
			note.Tags = append(note.Tags, tag)
		}
		if e := tx.Create(&note).Error; e != nil {
			return e
		}
		return tx.Preload("User").Preload("Tags").First(&note, note.ID).Error
	})
	if e != nil {
		failure(w, e)
		return
	}
	writeJSON(w, 201, note)
}
func (h *Handlers) GetNoteByID(w http.ResponseWriter, r *http.Request) {
	id, e := strconv.ParseUint(chi.URLParam(r, "id"), 10, 64)
	if e != nil || id == 0 {
		writeErr(w, 400, "bad id")
		return
	}
	var n models.Note
	if e := h.db.WithContext(r.Context()).Preload("User").Preload("Tags").First(&n, id).Error; e != nil {
		failure(w, e)
		return
	}
	writeJSON(w, 200, n)
}

func (h *Handlers) GetUsers(w http.ResponseWriter, r *http.Request) {

}
