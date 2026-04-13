package service

import (
	"context"
	"errors"
	"testing"
)

func Test_chunkCount(t *testing.T) {
	if chunkCount(0, 1024) != 0 {
		t.Fatal("zero size")
	}
	if chunkCount(1, 1024) != 1 {
		t.Fatal("single chunk")
	}
	if chunkCount(1024, 1024) != 1 {
		t.Fatal("exact")
	}
	if chunkCount(1025, 1024) != 2 {
		t.Fatal("spill")
	}
}

func Test_expectedChunkSize(t *testing.T) {
	if expectedChunkSize(1025, 1024, 0) != 1024 {
		t.Fatal("first full")
	}
	if expectedChunkSize(1025, 1024, 1) != 1 {
		t.Fatal("last partial")
	}
	if expectedChunkSize(2048, 1024, 1) != 1024 {
		t.Fatal("two equal chunks")
	}
}

func TestFileService_InitUpload_validationNoStorageHit(t *testing.T) {
	ctx := context.Background()
	s := &FileService{}
	_, _, _, err := s.InitUpload(ctx, "u1", InitUploadInput{
		FileName:  "",
		Size:      100,
		MimeType:  "application/pdf",
		ChunkSize: 256 * 1024,
	})
	if !errors.Is(err, ErrInvalidChunk) {
		t.Fatalf("empty fileName: %v", err)
	}

	_, _, _, err = s.InitUpload(ctx, "u1", InitUploadInput{
		FileName:  "a.pdf",
		Size:      0,
		MimeType:  "application/pdf",
		ChunkSize: 256 * 1024,
	})
	if !errors.Is(err, ErrInvalidChunk) {
		t.Fatalf("zero size: %v", err)
	}

	_, _, _, err = s.InitUpload(ctx, "u1", InitUploadInput{
		FileName:  "a.pdf",
		Size:      (100 << 20) + 1,
		MimeType:  "application/pdf",
		ChunkSize: 256 * 1024,
	})
	if !errors.Is(err, ErrFileSizeExceeded) {
		t.Fatalf("oversize: %v", err)
	}

	_, _, _, err = s.InitUpload(ctx, "u1", InitUploadInput{
		FileName:  "a.pdf",
		Size:      1000,
		MimeType:  "application/pdf",
		ChunkSize: 128 * 1024,
	})
	if !errors.Is(err, ErrInvalidChunk) {
		t.Fatalf("chunk too small: %v", err)
	}

	_, _, _, err = s.InitUpload(ctx, "u1", InitUploadInput{
		FileName:  "a.pdf",
		Size:      1000,
		MimeType:  "application/x-unknown",
		ChunkSize: 256 * 1024,
	})
	if !errors.Is(err, ErrFileTypeNotAllowed) {
		t.Fatalf("bad mime: %v", err)
	}

	_, _, _, err = s.InitUpload(ctx, "u1", InitUploadInput{
		FileName:  "a.pdf",
		Size:      1000,
		MimeType:  "application/pdf",
		ChunkSize: (8 << 20) + 1,
	})
	if !errors.Is(err, ErrInvalidChunk) {
		t.Fatalf("chunk too large: %v", err)
	}

	_, _, _, err = s.InitUpload(ctx, "u1", InitUploadInput{
		FileName:  "a.pdf",
		Size:      1000,
		MimeType:  "application/pdf",
		ChunkSize: 256 * 1024,
	})
	if err == nil || err.Error() != "file storage not configured" {
		t.Fatalf("expected storage not configured after validation, got %v", err)
	}
}
