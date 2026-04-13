package model

import "time"

// FileObject maps to `file_objects`.
type FileObject struct {
	ID          string    `gorm:"column:id;type:char(36);primaryKey"`
	SHA256      string    `gorm:"column:sha256;type:char(64);not null;uniqueIndex"`
	SizeBytes   uint64    `gorm:"column:size_bytes;not null"`
	MimeType    *string   `gorm:"column:mime_type;type:varchar(128)"`
	StoragePath string    `gorm:"column:storage_path;type:varchar(1024);not null"`
	CreatedAt   time.Time `gorm:"column:created_at"`
}

func (FileObject) TableName() string { return "file_objects" }

// UploadSession maps to `upload_sessions`.
type UploadSession struct {
	ID          string     `gorm:"column:id;type:char(36);primaryKey"`
	UserID      string     `gorm:"column:user_id;type:char(36);not null;index"`
	FileName    string     `gorm:"column:file_name;type:varchar(512);not null"`
	TotalSize   uint64     `gorm:"column:total_size;not null"`
	ChunkSize   uint32     `gorm:"column:chunk_size;not null"`
	MimeType    *string    `gorm:"column:mime_type;type:varchar(128)"`
	Status      string     `gorm:"column:status;type:varchar(32);not null"`
	MergedFileID *string   `gorm:"column:merged_file_id;type:char(36)"`
	ExpiresAt   *time.Time `gorm:"column:expires_at"`
	CreatedAt   time.Time  `gorm:"column:created_at"`
}

func (UploadSession) TableName() string { return "upload_sessions" }

// FileChunk maps to `file_chunks`.
type FileChunk struct {
	ID              string    `gorm:"column:id;type:char(36);primaryKey"`
	UploadSessionID string    `gorm:"column:upload_session_id;type:char(36);not null;index"`
	ChunkIndex      uint32    `gorm:"column:chunk_index;not null"`
	ChunkSHA256     string    `gorm:"column:chunk_sha256;type:char(64);not null"`
	StoragePath     string    `gorm:"column:storage_path;type:varchar(1024);not null"`
	CreatedAt       time.Time `gorm:"column:created_at"`
}

func (FileChunk) TableName() string { return "file_chunks" }

// FileReference maps to `file_references`.
type FileReference struct {
	ID              string    `gorm:"column:id;type:char(36);primaryKey"`
	FileObjectID    string    `gorm:"column:file_object_id;type:char(36);not null;index"`
	RefType         string    `gorm:"column:ref_type;type:varchar(64);not null"`
	RefID           string    `gorm:"column:ref_id;type:char(36);not null"`
	CreatedByUserID string    `gorm:"column:created_by_user_id;type:char(36);not null"`
	CreatedAt       time.Time `gorm:"column:created_at"`
}

func (FileReference) TableName() string { return "file_references" }
