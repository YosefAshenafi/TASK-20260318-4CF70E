package repository

import (
	"context"
	"time"

	"gorm.io/gorm"

	"pharmaops/api/internal/access"
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

// applyAccessibleFileScope limits file_objects to rows linked to a case in the principal's data scope
// or to an upload session merged by this user (unlinked uploads).
func (r *FileRepository) applyAccessibleFileScope(ctx context.Context, db *gorm.DB, p *access.Principal, userID string) *gorm.DB {
	uploadSub := r.db.WithContext(ctx).Model(&model.UploadSession{}).
		Select("merged_file_id").
		Where("user_id = ? AND merged_file_id IS NOT NULL", userID)
	scopeExpr, scopeArgs, ok := buildDataScopeExpr(p, "c.institution_id", "c.department_id", "c.team_id")
	if !ok {
		return db.Where("file_objects.id IN (?)", uploadSub)
	}
	caseSub := r.db.WithContext(ctx).Model(&model.FileObject{}).
		Select("DISTINCT file_objects.id").
		Joins("INNER JOIN file_references fr ON fr.file_object_id = file_objects.id").
		Joins("INNER JOIN cases c ON fr.ref_type = 'case' AND fr.ref_id = c.id").
		Where(scopeExpr, scopeArgs...)
	return db.Where("(file_objects.id IN (?) OR file_objects.id IN (?))", caseSub, uploadSub)
}

// ListAccessibleFileObjects returns file metadata visible to the caller (case-linked within scope or own uploads).
func (r *FileRepository) ListAccessibleFileObjects(ctx context.Context, p *access.Principal, userID string, offset, limit int) ([]model.FileObject, int64, error) {
	base := r.db.WithContext(ctx).Model(&model.FileObject{})
	base = r.applyAccessibleFileScope(ctx, base, p, userID)
	var total int64
	if err := base.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var rows []model.FileObject
	err := r.db.WithContext(ctx).Model(&model.FileObject{}).
		Scopes(func(db *gorm.DB) *gorm.DB {
			return r.applyAccessibleFileScope(ctx, db, p, userID)
		}).
		Order("created_at DESC").
		Offset(offset).Limit(limit).
		Find(&rows).Error
	if err != nil {
		return nil, 0, err
	}
	return rows, total, nil
}

// IsFileObjectAccessible reports whether the file is linked to an in-scope case or merged from the user's upload session.
func (r *FileRepository) IsFileObjectAccessible(ctx context.Context, p *access.Principal, userID, fileID string) (bool, error) {
	var n int64
	err := r.db.WithContext(ctx).Model(&model.UploadSession{}).
		Where("merged_file_id = ? AND user_id = ?", fileID, userID).
		Count(&n).Error
	if err != nil {
		return false, err
	}
	if n > 0 {
		return true, nil
	}
	scopeExpr, scopeArgs, ok := buildDataScopeExpr(p, "c.institution_id", "c.department_id", "c.team_id")
	if !ok {
		return false, nil
	}
	err = r.db.WithContext(ctx).Model(&model.FileObject{}).
		Joins("INNER JOIN file_references fr ON fr.file_object_id = file_objects.id").
		Joins("INNER JOIN cases c ON fr.ref_type = 'case' AND fr.ref_id = c.id").
		Where("file_objects.id = ?", fileID).
		Where(scopeExpr, scopeArgs...).
		Count(&n).Error
	if err != nil {
		return false, err
	}
	return n > 0, nil
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

// CreateCaseAttachmentIndex inserts a row into case_attachment_indexes.
func (r *FileRepository) CreateCaseAttachmentIndex(ctx context.Context, idx *model.CaseAttachmentIndex) error {
	return r.db.WithContext(ctx).Create(idx).Error
}

func (r *FileRepository) ListCaseAttachmentIndexes(ctx context.Context, caseID string) ([]model.CaseAttachmentIndex, error) {
	var rows []model.CaseAttachmentIndex
	err := r.db.WithContext(ctx).
		Where("case_id = ?", caseID).
		Order("created_at ASC").
		Find(&rows).Error
	return rows, err
}

func (r *FileRepository) DeleteCaseAttachmentIndexByCaseAndFile(ctx context.Context, caseID, fileID string) error {
	return r.db.WithContext(ctx).
		Where("case_id = ? AND file_object_id = ?", caseID, fileID).
		Delete(&model.CaseAttachmentIndex{}).Error
}

func (r *FileRepository) DeleteCaseFileReference(ctx context.Context, caseID, fileID string) error {
	return r.db.WithContext(ctx).
		Where("ref_type = ? AND ref_id = ? AND file_object_id = ?", "case", caseID, fileID).
		Delete(&model.FileReference{}).Error
}

// FileObjectExists returns true if a file_object row with the given ID exists.
func (r *FileRepository) FileObjectExists(ctx context.Context, id string) bool {
	var count int64
	r.db.WithContext(ctx).Model(&model.FileObject{}).Where("id = ?", id).Count(&count)
	return count > 0
}
