package repository

import (
	"context"
	"time"

	"gorm.io/gorm"

	"pharmaops/api/internal/model"
)

type FileRepository struct {
	db *gorm.DB
}

func NewFileRepository(db *gorm.DB) *FileRepository {
	return &FileRepository{db: db}
}

// GetDB exposes the DB for transactions in the file service.
func (r *FileRepository) GetDB() *gorm.DB {
	return r.db
}

func (r *FileRepository) CreateUploadSession(ctx context.Context, row *model.UploadSession) error {
	return r.db.WithContext(ctx).Create(row).Error
}

func (r *FileRepository) GetUploadSession(ctx context.Context, id string) (*model.UploadSession, error) {
	var row model.UploadSession
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&row).Error
	if err != nil {
		return nil, err
	}
	return &row, nil
}

func (r *FileRepository) GetUploadSessionForUser(ctx context.Context, id, userID string) (*model.UploadSession, error) {
	var row model.UploadSession
	err := r.db.WithContext(ctx).Where("id = ? AND user_id = ?", id, userID).First(&row).Error
	if err != nil {
		return nil, err
	}
	return &row, nil
}

func (r *FileRepository) CountChunks(ctx context.Context, uploadSessionID string) (int64, error) {
	var n int64
	err := r.db.WithContext(ctx).Model(&model.FileChunk{}).Where("upload_session_id = ?", uploadSessionID).Count(&n).Error
	return n, err
}

func (r *FileRepository) ListChunkIndices(ctx context.Context, uploadSessionID string) ([]uint32, error) {
	var indices []uint32
	err := r.db.WithContext(ctx).Model(&model.FileChunk{}).
		Where("upload_session_id = ?", uploadSessionID).
		Order("chunk_index ASC").
		Pluck("chunk_index", &indices).Error
	return indices, err
}

func (r *FileRepository) ListChunksOrdered(ctx context.Context, uploadSessionID string) ([]model.FileChunk, error) {
	var rows []model.FileChunk
	err := r.db.WithContext(ctx).Where("upload_session_id = ?", uploadSessionID).
		Order("chunk_index ASC").Find(&rows).Error
	return rows, err
}

func (r *FileRepository) DeleteChunkByIndex(ctx context.Context, uploadSessionID string, chunkIndex uint32) error {
	res := r.db.WithContext(ctx).Where("upload_session_id = ? AND chunk_index = ?", uploadSessionID, chunkIndex).
		Delete(&model.FileChunk{})
	if res.Error != nil {
		return res.Error
	}
	return nil
}

func (r *FileRepository) CreateChunk(ctx context.Context, row *model.FileChunk) error {
	return r.db.WithContext(ctx).Create(row).Error
}

func (r *FileRepository) GetFileObject(ctx context.Context, id string) (*model.FileObject, error) {
	var row model.FileObject
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&row).Error
	if err != nil {
		return nil, err
	}
	return &row, nil
}

func (r *FileRepository) GetFileObjectBySHA256(ctx context.Context, sha256 string) (*model.FileObject, error) {
	var row model.FileObject
	err := r.db.WithContext(ctx).Where("sha256 = ?", sha256).First(&row).Error
	if err != nil {
		return nil, err
	}
	return &row, nil
}

func (r *FileRepository) CreateFileObject(ctx context.Context, row *model.FileObject) error {
	return r.db.WithContext(ctx).Create(row).Error
}

func (r *FileRepository) InsertDedupIndex(ctx context.Context, sha256, fileObjectID string) error {
	return r.db.WithContext(ctx).Exec(
		`INSERT INTO file_dedup_index (sha256, file_object_id) VALUES (?, ?)`,
		sha256, fileObjectID,
	).Error
}

func (r *FileRepository) UpdateUploadSessionMerged(ctx context.Context, id, status string, mergedFileID string) error {
	return r.db.WithContext(ctx).Model(&model.UploadSession{}).Where("id = ?", id).Updates(map[string]any{
		"status":         status,
		"merged_file_id": mergedFileID,
	}).Error
}

func (r *FileRepository) DeleteChunksForSession(ctx context.Context, uploadSessionID string) error {
	return r.db.WithContext(ctx).Where("upload_session_id = ?", uploadSessionID).Delete(&model.FileChunk{}).Error
}

func (r *FileRepository) ListFileObjects(ctx context.Context, offset, limit int) ([]model.FileObject, int64, error) {
	var total int64
	if err := r.db.WithContext(ctx).Model(&model.FileObject{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var rows []model.FileObject
	err := r.db.WithContext(ctx).Order("created_at DESC").Offset(offset).Limit(limit).Find(&rows).Error
	if err != nil {
		return nil, 0, err
	}
	return rows, total, nil
}

func (r *FileRepository) CreateFileReference(ctx context.Context, row *model.FileReference) error {
	return r.db.WithContext(ctx).Create(row).Error
}

// ExpireStaleUploads marks old initialized sessions as failed (best-effort maintenance hook).
func (r *FileRepository) ExpireStaleUploads(ctx context.Context, before time.Time) error {
	return r.db.WithContext(ctx).Model(&model.UploadSession{}).
		Where("status = ? AND expires_at IS NOT NULL AND expires_at < ?", "initialized", before).
		Update("status", "failed").Error
}
