package service

import (
	"bytes"
	"context"
	"errors"
	"testing"
	"time"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"pharmaops/api/internal/model"
	"pharmaops/api/internal/repository"
)

func TestFileService_CompleteUpload_deduplicatesIdenticalContent(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	if err := db.AutoMigrate(&model.User{}, &model.UploadSession{}, &model.FileChunk{}, &model.FileObject{}); err != nil {
		t.Fatal(err)
	}
	if err := db.Exec(`CREATE TABLE file_dedup_index (sha256 TEXT PRIMARY KEY, file_object_id TEXT NOT NULL)`).Error; err != nil {
		t.Fatal(err)
	}
	if err := db.Create(&model.User{
		ID:           "u1",
		Username:     "u1",
		PasswordHash: "x",
		DisplayName:  "U1",
		IsActive:     true,
		CreatedAt:    time.Now().UTC(),
		UpdatedAt:    time.Now().UTC(),
	}).Error; err != nil {
		t.Fatal(err)
	}

	root := t.TempDir()
	files := repository.NewFileRepository(db)
	cases := repository.NewCaseRepository(db)
	svc := NewFileService(root, files, cases, NewAuditService(nil))
	payload := bytes.Repeat([]byte("resume-data-"), 30000)
	ctx := context.Background()

	uploadAndComplete := func() (*CompleteUploadResponse, error) {
		uploadID, totalChunks, _, err := svc.InitUpload(ctx, "u1", InitUploadInput{
			FileName:  "resume.txt",
			Size:      uint64(len(payload)),
			MimeType:  "text/plain",
			ChunkSize: 256 * 1024,
		})
		if err != nil {
			return nil, err
		}
		for i := uint64(0); i < totalChunks; i++ {
			start := i * 256 * 1024
			end := start + 256*1024
			if end > uint64(len(payload)) {
				end = uint64(len(payload))
			}
			if err := svc.PutChunk(ctx, "u1", uploadID, uint32(i), payload[start:end]); err != nil {
				return nil, err
			}
		}
		return svc.CompleteUpload(ctx, "u1", uploadID, CompleteUploadInput{}, AuditRequestMeta{OperatorUserID: "u1"})
	}

	first, err := uploadAndComplete()
	if err != nil {
		t.Fatal(err)
	}
	if first.Deduplicated {
		t.Fatal("first upload should create a new file object")
	}

	second, err := uploadAndComplete()
	if err != nil {
		t.Fatal(err)
	}
	if !second.Deduplicated {
		t.Fatal("second upload should be deduplicated")
	}
	if first.FileID != second.FileID {
		t.Fatalf("expected same file id from dedup, got %s and %s", first.FileID, second.FileID)
	}
}

func TestFileService_CompleteUpload_rejectsMimeSpoofedPayload(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	if err := db.AutoMigrate(&model.User{}, &model.UploadSession{}, &model.FileChunk{}, &model.FileObject{}); err != nil {
		t.Fatal(err)
	}
	if err := db.Exec(`CREATE TABLE file_dedup_index (sha256 TEXT PRIMARY KEY, file_object_id TEXT NOT NULL)`).Error; err != nil {
		t.Fatal(err)
	}
	if err := db.Create(&model.User{
		ID:           "u1",
		Username:     "u1",
		PasswordHash: "x",
		DisplayName:  "U1",
		IsActive:     true,
		CreatedAt:    time.Now().UTC(),
		UpdatedAt:    time.Now().UTC(),
	}).Error; err != nil {
		t.Fatal(err)
	}

	root := t.TempDir()
	files := repository.NewFileRepository(db)
	cases := repository.NewCaseRepository(db)
	svc := NewFileService(root, files, cases, NewAuditService(nil))

	payload := bytes.Repeat([]byte("not-a-real-pdf"), 20000)
	ctx := context.Background()
	uploadID, totalChunks, _, err := svc.InitUpload(ctx, "u1", InitUploadInput{
		FileName:  "resume.pdf",
		Size:      uint64(len(payload)),
		MimeType:  "application/pdf",
		ChunkSize: 256 * 1024,
	})
	if err != nil {
		t.Fatal(err)
	}
	for i := uint64(0); i < totalChunks; i++ {
		start := i * 256 * 1024
		end := start + 256*1024
		if end > uint64(len(payload)) {
			end = uint64(len(payload))
		}
		if err := svc.PutChunk(ctx, "u1", uploadID, uint32(i), payload[start:end]); err != nil {
			t.Fatal(err)
		}
	}

	_, err = svc.CompleteUpload(ctx, "u1", uploadID, CompleteUploadInput{}, AuditRequestMeta{OperatorUserID: "u1"})
	if !errors.Is(err, ErrFileTypeNotAllowed) {
		t.Fatalf("expected ErrFileTypeNotAllowed for spoofed payload, got %v", err)
	}
}
