package service

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/google/uuid"
	"pharmaops/api/internal/access"
	"pharmaops/api/internal/model"
	"pharmaops/api/internal/repository"
)

const (
	maxFileBytes        = 100 << 20 // 100 MiB
	minChunkBytes       = 256 * 1024
	maxChunkBytes       = 8 << 20
	uploadSessionTTL    = 24 * time.Hour
	chunkFilePerm       = 0o600
	dirPerm             = 0o700
)

var allowedMimeTypes = map[string]struct{}{
	"application/pdf": {},
	"image/jpeg":      {},
	"image/png":       {},
	"image/webp":      {},
	"text/plain":      {},
	"application/vnd.openxmlformats-officedocument.wordprocessingml.document": {},
	"application/msword": {},
}

// File errors.
var (
	ErrFileNotFound          = errors.New("file not found")
	ErrFileTypeNotAllowed    = errors.New("file type not allowed")
	ErrFileSizeExceeded      = errors.New("file size exceeded")
	ErrFileChunkMissing      = errors.New("chunk missing")
	ErrFileHashMismatch      = errors.New("hash mismatch")
	ErrUploadNotFound        = errors.New("upload not found")
	ErrUploadExpired         = errors.New("upload expired")
	ErrUploadAlreadyComplete = errors.New("upload already complete")
	ErrInvalidChunk          = errors.New("invalid chunk")
)

type FileService struct {
	root     string
	files    *repository.FileRepository
	cases    *repository.CaseRepository
	audit    *AuditService
}

func NewFileService(root string, files *repository.FileRepository, cases *repository.CaseRepository, audit *AuditService) *FileService {
	return &FileService{root: root, files: files, cases: cases, audit: audit}
}

// InitUploadInput matches POST /files/uploads/init.
type InitUploadInput struct {
	FileName  string
	Size      uint64
	MimeType  string
	ChunkSize uint32
}

func (s *FileService) InitUpload(ctx context.Context, userID string, in InitUploadInput) (uploadID string, totalChunks uint64, expiresAt time.Time, err error) {
	if in.FileName == "" || in.Size == 0 || in.ChunkSize == 0 {
		return "", 0, time.Time{}, ErrInvalidChunk
	}
	if in.Size > maxFileBytes {
		return "", 0, time.Time{}, ErrFileSizeExceeded
	}
	if in.ChunkSize < minChunkBytes || in.ChunkSize > maxChunkBytes {
		return "", 0, time.Time{}, ErrInvalidChunk
	}
	if _, ok := allowedMimeTypes[in.MimeType]; !ok {
		return "", 0, time.Time{}, ErrFileTypeNotAllowed
	}
	total := (in.Size + uint64(in.ChunkSize) - 1) / uint64(in.ChunkSize)
	if total == 0 || total > 100000 {
		return "", 0, time.Time{}, ErrInvalidChunk
	}
	if s.root == "" {
		return "", 0, time.Time{}, errors.New("file storage not configured")
	}
	id := uuid.NewString()
	exp := time.Now().UTC().Add(uploadSessionTTL)
	mime := in.MimeType
	row := &model.UploadSession{
		ID:          id,
		UserID:      userID,
		FileName:    in.FileName,
		TotalSize:   in.Size,
		ChunkSize:   in.ChunkSize,
		MimeType:    &mime,
		Status:      "initialized",
		MergedFileID: nil,
		ExpiresAt:   &exp,
		CreatedAt:   time.Now().UTC(),
	}
	if err := s.files.CreateUploadSession(ctx, row); err != nil {
		return "", 0, time.Time{}, err
	}
	chunkDir := filepath.Join(s.root, "chunks", id)
	if err := os.MkdirAll(chunkDir, dirPerm); err != nil {
		return "", 0, time.Time{}, err
	}
	return id, total, exp, nil
}

// PutChunk writes one chunk and records metadata.
func (s *FileService) PutChunk(ctx context.Context, userID, uploadID string, chunkIndex uint32, body []byte) error {
	if s.root == "" {
		return errors.New("file storage not configured")
	}
	sess, err := s.files.GetUploadSessionForUser(ctx, uploadID, userID)
	if repository.IsNotFound(err) {
		return ErrUploadNotFound
	}
	if err != nil {
		return err
	}
	if sess.Status != "initialized" {
		return ErrUploadAlreadyComplete
	}
	if sess.ExpiresAt != nil && time.Now().UTC().After(*sess.ExpiresAt) {
		return ErrUploadExpired
	}
	totalChunks := chunkCount(sess.TotalSize, uint64(sess.ChunkSize))
	if uint64(chunkIndex) >= totalChunks {
		return ErrInvalidChunk
	}
	expected := expectedChunkSize(sess.TotalSize, uint64(sess.ChunkSize), chunkIndex)
	if uint64(len(body)) != expected {
		return ErrInvalidChunk
	}
	h := sha256.Sum256(body)
	chunkHash := hex.EncodeToString(h[:])
	relPath := filepath.Join("chunks", uploadID, fmt.Sprintf("%d.part", chunkIndex))
	absPath := filepath.Join(s.root, relPath)
	if err := os.MkdirAll(filepath.Dir(absPath), dirPerm); err != nil {
		return err
	}
	if err := os.WriteFile(absPath, body, chunkFilePerm); err != nil {
		return err
	}
	_ = s.files.DeleteChunkByIndex(ctx, uploadID, chunkIndex)
	ch := &model.FileChunk{
		ID:              uuid.NewString(),
		UploadSessionID: uploadID,
		ChunkIndex:      chunkIndex,
		ChunkSHA256:     chunkHash,
		StoragePath:     filepath.ToSlash(relPath),
		CreatedAt:       time.Now().UTC(),
	}
	if err := s.files.CreateChunk(ctx, ch); err != nil {
		return err
	}
	return nil
}

func chunkCount(totalSize, chunkSize uint64) uint64 {
	if totalSize == 0 { return 0 }
	return (totalSize + chunkSize - 1) / chunkSize
}

func expectedChunkSize(totalSize, chunkSize uint64, chunkIndex uint32) uint64 {
	n := chunkCount(totalSize, chunkSize)
	last := n - 1
	if uint64(chunkIndex) < last {
		return chunkSize
	}
	rem := totalSize % chunkSize
	if rem == 0 {
		return chunkSize
	}
	return rem
}

// CompleteUploadInput optional client hash for verification.
type CompleteUploadInput struct {
	SHA256 *string
}

// CompleteUploadResponse matches api-spec completion data.
type CompleteUploadResponse struct {
	FileID         string `json:"fileId"`
	SHA256         string `json:"sha256"`
	Deduplicated   bool   `json:"deduplicated"`
}

func (s *FileService) CompleteUpload(ctx context.Context, userID, uploadID string, in CompleteUploadInput, meta AuditRequestMeta) (*CompleteUploadResponse, error) {
	if s.root == "" {
		return nil, errors.New("file storage not configured")
	}
	sess, err := s.files.GetUploadSessionForUser(ctx, uploadID, userID)
	if repository.IsNotFound(err) {
		return nil, ErrUploadNotFound
	}
	if err != nil {
		return nil, err
	}
	if sess.Status != "initialized" {
		return nil, ErrUploadAlreadyComplete
	}
	if sess.ExpiresAt != nil && time.Now().UTC().After(*sess.ExpiresAt) {
		return nil, ErrUploadExpired
	}
	chunks, err := s.files.ListChunksOrdered(ctx, uploadID)
	if err != nil {
		return nil, err
	}
	totalChunks := chunkCount(sess.TotalSize, uint64(sess.ChunkSize))
	if uint64(len(chunks)) != totalChunks {
		return nil, ErrFileChunkMissing
	}
	idxSeen := make(map[uint32]struct{}, len(chunks))
	for _, c := range chunks {
		idxSeen[c.ChunkIndex] = struct{}{}
	}
	for i := uint64(0); i < totalChunks; i++ {
		if _, ok := idxSeen[uint32(i)]; !ok {
			return nil, ErrFileChunkMissing
		}
	}
	mergedRel := filepath.Join("tmp", uploadID+"_merged.bin")
	mergedAbs := filepath.Join(s.root, mergedRel)
	if err := os.MkdirAll(filepath.Dir(mergedAbs), dirPerm); err != nil {
		return nil, err
	}
	out, err := os.Create(mergedAbs)
	if err != nil {
		return nil, err
	}
	hash := sha256.New()
	mw := io.MultiWriter(out, hash)
	for _, c := range chunks {
		p := filepath.Join(s.root, filepath.FromSlash(c.StoragePath))
		f, err := os.Open(p)
		if err != nil {
			out.Close()
			os.Remove(mergedAbs)
			return nil, err
		}
		_, err = io.Copy(mw, f)
		f.Close()
		if err != nil {
			out.Close()
			os.Remove(mergedAbs)
			return nil, err
		}
	}
	if err := out.Close(); err != nil {
		os.Remove(mergedAbs)
		return nil, err
	}
	finalHash := hex.EncodeToString(hash.Sum(nil))
	if in.SHA256 != nil && *in.SHA256 != "" && *in.SHA256 != finalHash {
		os.Remove(mergedAbs)
		return nil, ErrFileHashMismatch
	}
	// Deduplicate
	existing, err := s.files.GetFileObjectBySHA256(ctx, finalHash)
	if err == nil && existing != nil {
		_ = os.Remove(mergedAbs)
		_ = s.removeChunkDir(uploadID)
		_ = s.files.DeleteChunksForSession(ctx, uploadID)
		mid := existing.ID
		if err := s.files.UpdateUploadSessionMerged(ctx, uploadID, "completed", mid); err != nil {
			return nil, err
		}
		m := meta
		if m.OperatorUserID == "" {
			m.OperatorUserID = userID
		}
		_ = s.audit.LogMutation(ctx, AuditMutationInput{
			Module:     "files",
			Operation:  "file.upload_complete",
			TargetType: "file_object",
			TargetID:   existing.ID,
			After: map[string]any{
				"sha256":       finalHash,
				"deduplicated": true,
				"uploadId":     uploadID,
				"sizeBytes":    sess.TotalSize,
			},
			Meta: m,
		})
		return &CompleteUploadResponse{FileID: existing.ID, SHA256: finalHash, Deduplicated: true}, nil
	}
	if err != nil && !repository.IsNotFound(err) {
		os.Remove(mergedAbs)
		return nil, err
	}
	fileID := uuid.NewString()
	objRel := filepath.Join("objects", fileID+".bin")
	objAbs := filepath.Join(s.root, objRel)
	if err := os.MkdirAll(filepath.Dir(objAbs), dirPerm); err != nil {
		os.Remove(mergedAbs)
		return nil, err
	}
	if err := os.Rename(mergedAbs, objAbs); err != nil {
		os.Remove(mergedAbs)
		return nil, err
	}
	mime := ""
	if sess.MimeType != nil {
		mime = *sess.MimeType
	}
	var mimePtr *string
	if mime != "" {
		mimePtr = &mime
	}
	fo := &model.FileObject{
		ID:          fileID,
		SHA256:      finalHash,
		SizeBytes:   sess.TotalSize,
		MimeType:    mimePtr,
		StoragePath: filepath.ToSlash(objRel),
		CreatedAt:   time.Now().UTC(),
	}
	tx := s.files.GetDB().WithContext(ctx).Begin()
	if err := tx.Error; err != nil {
		return nil, err
	}
	if err := tx.Create(fo).Error; err != nil {
		tx.Rollback()
		os.Remove(objAbs)
		return nil, err
	}
	if err := tx.Exec(`INSERT INTO file_dedup_index (sha256, file_object_id) VALUES (?, ?)`, finalHash, fileID).Error; err != nil {
		tx.Rollback()
		os.Remove(objAbs)
		return nil, err
	}
	if err := tx.Model(&model.UploadSession{}).Where("id = ?", uploadID).Updates(map[string]any{
		"status":         "completed",
		"merged_file_id": fileID,
	}).Error; err != nil {
		tx.Rollback()
		os.Remove(objAbs)
		return nil, err
	}
	if err := tx.Where("upload_session_id = ?", uploadID).Delete(&model.FileChunk{}).Error; err != nil {
		tx.Rollback()
		return nil, err
	}
	if err := tx.Commit().Error; err != nil {
		return nil, err
	}
	_ = s.removeChunkDir(uploadID)
	m := meta
	if m.OperatorUserID == "" {
		m.OperatorUserID = userID
	}
	_ = s.audit.LogMutation(ctx, AuditMutationInput{
		Module:     "files",
		Operation:  "file.upload_complete",
		TargetType: "file_object",
		TargetID:   fileID,
		After: map[string]any{
			"sha256":       finalHash,
			"deduplicated": false,
			"uploadId":     uploadID,
			"sizeBytes":    sess.TotalSize,
		},
		Meta: m,
	})
	return &CompleteUploadResponse{FileID: fileID, SHA256: finalHash, Deduplicated: false}, nil
}

func (s *FileService) removeChunkDir(uploadID string) error {
	d := filepath.Join(s.root, "chunks", uploadID)
	return os.RemoveAll(d)
}

// UploadSessionDTO for GET /files/uploads/{id}.
type UploadSessionDTO struct {
	ID              string  `json:"id"`
	Status          string  `json:"status"`
	FileName        string  `json:"fileName"`
	TotalSize       uint64  `json:"totalSize"`
	ChunkSize       uint32  `json:"chunkSize"`
	MimeType        *string `json:"mimeType,omitempty"`
	ReceivedChunks  int64   `json:"receivedChunks"`
	TotalChunks     uint64  `json:"totalChunks"`
	MergedFileID    *string `json:"mergedFileId,omitempty"`
	ExpiresAt       *string `json:"expiresAt,omitempty"`
}

func (s *FileService) GetUploadSession(ctx context.Context, userID, uploadID string) (*UploadSessionDTO, error) {
	sess, err := s.files.GetUploadSessionForUser(ctx, uploadID, userID)
	if repository.IsNotFound(err) {
		return nil, ErrUploadNotFound
	}
	if err != nil {
		return nil, err
	}
	n, err := s.files.CountChunks(ctx, uploadID)
	if err != nil {
		return nil, err
	}
	totalChunks := chunkCount(sess.TotalSize, uint64(sess.ChunkSize))
	var exp *string
	if sess.ExpiresAt != nil {
		t := sess.ExpiresAt.UTC().Format(time.RFC3339Nano)
		exp = &t
	}
	return &UploadSessionDTO{
		ID:             sess.ID,
		Status:         sess.Status,
		FileName:       sess.FileName,
		TotalSize:      sess.TotalSize,
		ChunkSize:      sess.ChunkSize,
		MimeType:       sess.MimeType,
		ReceivedChunks: n,
		TotalChunks:    totalChunks,
		MergedFileID:   sess.MergedFileID,
		ExpiresAt:      exp,
	}, nil
}

// FileObjectDTO for GET /files/{id} and list.
type FileObjectDTO struct {
	ID        string  `json:"id"`
	SHA256    string  `json:"sha256"`
	SizeBytes uint64  `json:"sizeBytes"`
	MimeType  *string `json:"mimeType,omitempty"`
	CreatedAt string  `json:"createdAt"`
}

func (s *FileService) GetFile(ctx context.Context, pr *access.Principal, userID, fileID string) (*FileObjectDTO, error) {
	ok, err := s.files.IsFileObjectAccessible(ctx, pr, userID, fileID)
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, ErrFileNotFound
	}
	fo, err := s.files.GetFileObject(ctx, fileID)
	if repository.IsNotFound(err) {
		return nil, ErrFileNotFound
	}
	if err != nil {
		return nil, err
	}
	return fileToDTO(fo), nil
}

func fileToDTO(fo *model.FileObject) *FileObjectDTO {
	return &FileObjectDTO{
		ID:        fo.ID,
		SHA256:    fo.SHA256,
		SizeBytes: fo.SizeBytes,
		MimeType:  fo.MimeType,
		CreatedAt: fo.CreatedAt.UTC().Format(time.RFC3339Nano),
	}
}

func (s *FileService) ListFiles(ctx context.Context, pr *access.Principal, userID string, offset, limit int) ([]FileObjectDTO, int64, error) {
	rows, total, err := s.files.ListAccessibleFileObjects(ctx, pr, userID, offset, limit)
	if err != nil {
		return nil, 0, err
	}
	out := make([]FileObjectDTO, 0, len(rows))
	for i := range rows {
		out = append(out, *fileToDTO(&rows[i]))
	}
	return out, total, nil
}

// ResolvedObjectPath returns absolute filesystem path for a stored file.
func (s *FileService) ResolvedObjectPath(fo *model.FileObject) string {
	return filepath.Join(s.root, filepath.FromSlash(fo.StoragePath))
}

func (s *FileService) GetFileObject(ctx context.Context, pr *access.Principal, userID, fileID string) (*model.FileObject, error) {
	ok, err := s.files.IsFileObjectAccessible(ctx, pr, userID, fileID)
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, ErrFileNotFound
	}
	fo, err := s.files.GetFileObject(ctx, fileID)
	if repository.IsNotFound(err) {
		return nil, ErrFileNotFound
	}
	return fo, err
}

// LinkFileInput for POST /files/{id}/link.
type LinkFileInput struct {
	RefType string
	RefID   string
}

func (s *FileService) LinkFile(ctx context.Context, userID string, pr *access.Principal, fileID string, in LinkFileInput, meta AuditRequestMeta) error {
	if pr == nil || userID == "" {
		return ErrForbiddenScope
	}
	_, err := s.GetFileObject(ctx, pr, userID, fileID)
	if errors.Is(err, ErrFileNotFound) {
		return ErrFileNotFound
	}
	if err != nil {
		return err
	}
	if in.RefType == "case" {
		if err := requireScope(pr); err != nil {
			return err
		}
		_, err := s.cases.GetCase(ctx, in.RefID, pr)
		if repository.IsNotFound(err) {
			return ErrFileNotFound
		}
		if err != nil {
			return err
		}
	} else {
		return errors.New("unsupported ref type")
	}
	ref := &model.FileReference{
		ID:              uuid.NewString(),
		FileObjectID:    fileID,
		RefType:         in.RefType,
		RefID:           in.RefID,
		CreatedByUserID: userID,
		CreatedAt:       time.Now().UTC(),
	}
	if err := s.files.CreateFileReference(ctx, ref); err != nil {
		return err
	}
	m := meta
	if m.OperatorUserID == "" {
		m.OperatorUserID = userID
	}
	_ = s.audit.LogMutation(ctx, AuditMutationInput{
		Module:     "files",
		Operation:  "file.link",
		TargetType: "file_object",
		TargetID:   fileID,
		After: map[string]any{
			"referenceId": ref.ID,
			"refType":     in.RefType,
			"refId":       in.RefID,
		},
		Meta: m,
	})
	return nil
}
