package handler

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"

	"pharmaops/api/internal/middleware"
	"pharmaops/api/internal/response"
	"pharmaops/api/internal/service"
)

type FileHandler struct {
	svc *service.FileService
}

func NewFileHandler(svc *service.FileService) *FileHandler {
	return &FileHandler{svc: svc}
}

type initUploadBody struct {
	FileName  string `json:"fileName" binding:"required"`
	Size      uint64 `json:"size" binding:"required"`
	MimeType  string `json:"mimeType" binding:"required"`
	ChunkSize uint32 `json:"chunkSize" binding:"required"`
}

func (h *FileHandler) InitUpload(c *gin.Context) {
	uid := c.GetString("userID")
	if uid == "" {
		response.Error(c, http.StatusUnauthorized, "AUTH_SESSION_EXPIRED", "missing user")
		return
	}
	var body initUploadBody
	if err := c.ShouldBindJSON(&body); err != nil {
		response.Error(c, http.StatusBadRequest, "VALIDATION_ERROR", "invalid request body")
		return
	}
	uploadID, totalChunks, exp, err := h.svc.InitUpload(c.Request.Context(), uid, service.InitUploadInput{
		FileName:  body.FileName,
		Size:      body.Size,
		MimeType:  body.MimeType,
		ChunkSize: body.ChunkSize,
	})
	if errors.Is(err, service.ErrFileSizeExceeded) {
		response.Error(c, http.StatusBadRequest, "FILE_SIZE_EXCEEDED", "file too large")
		return
	}
	if errors.Is(err, service.ErrFileTypeNotAllowed) {
		response.Error(c, http.StatusBadRequest, "FILE_TYPE_NOT_ALLOWED", "mime type not allowed")
		return
	}
	if errors.Is(err, service.ErrInvalidChunk) {
		response.Error(c, http.StatusBadRequest, "VALIDATION_ERROR", "invalid chunk configuration")
		return
	}
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to init upload")
		return
	}
	response.OK(c, gin.H{
		"uploadId":    uploadID,
		"totalChunks": totalChunks,
		"expiresAt":   exp.UTC().Format(time.RFC3339Nano),
	})
}

func (h *FileHandler) PutChunk(c *gin.Context) {
	uid := c.GetString("userID")
	if uid == "" {
		response.Error(c, http.StatusUnauthorized, "AUTH_SESSION_EXPIRED", "missing user")
		return
	}
	uploadID := c.Param("uploadId")
	idxStr := c.Param("chunkIndex")
	u64, err := strconv.ParseUint(idxStr, 10, 32)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "VALIDATION_ERROR", "invalid chunk index")
		return
	}
	chunkIndex := uint32(u64)
	// Hard cap per request body (matches max chunk size in service).
	const maxRead = 8<<20 + 1024
	data, err := io.ReadAll(io.LimitReader(c.Request.Body, maxRead))
	if err != nil {
		response.Error(c, http.StatusBadRequest, "VALIDATION_ERROR", "failed to read body")
		return
	}
	err = h.svc.PutChunk(c.Request.Context(), uid, uploadID, chunkIndex, data)
	if errors.Is(err, service.ErrUploadNotFound) {
		response.Error(c, http.StatusNotFound, "FILE_NOT_FOUND", "upload not found")
		return
	}
	if errors.Is(err, service.ErrUploadExpired) {
		response.Error(c, http.StatusGone, "FILE_NOT_FOUND", "upload expired")
		return
	}
	if errors.Is(err, service.ErrUploadAlreadyComplete) {
		response.Error(c, http.StatusConflict, "VALIDATION_ERROR", "upload already completed")
		return
	}
	if errors.Is(err, service.ErrInvalidChunk) {
		response.Error(c, http.StatusBadRequest, "VALIDATION_ERROR", "invalid chunk")
		return
	}
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to store chunk")
		return
	}
	c.Status(http.StatusNoContent)
}

type completeUploadBody struct {
	SHA256 *string `json:"sha256"`
}

func (h *FileHandler) CompleteUpload(c *gin.Context) {
	uid := c.GetString("userID")
	if uid == "" {
		response.Error(c, http.StatusUnauthorized, "AUTH_SESSION_EXPIRED", "missing user")
		return
	}
	uploadID := c.Param("uploadId")
	var body completeUploadBody
	if raw, _ := io.ReadAll(c.Request.Body); len(bytes.TrimSpace(raw)) > 0 {
		_ = json.Unmarshal(raw, &body)
	}
	out, err := h.svc.CompleteUpload(c.Request.Context(), uid, uploadID, service.CompleteUploadInput{SHA256: body.SHA256}, auditRequestMeta(c))
	if errors.Is(err, service.ErrUploadNotFound) {
		response.Error(c, http.StatusNotFound, "FILE_NOT_FOUND", "upload not found")
		return
	}
	if errors.Is(err, service.ErrUploadExpired) {
		response.Error(c, http.StatusGone, "FILE_NOT_FOUND", "upload expired")
		return
	}
	if errors.Is(err, service.ErrUploadAlreadyComplete) {
		response.Error(c, http.StatusConflict, "VALIDATION_ERROR", "upload already completed")
		return
	}
	if errors.Is(err, service.ErrFileChunkMissing) {
		response.Error(c, http.StatusBadRequest, "FILE_CHUNK_MISSING", "missing chunks")
		return
	}
	if errors.Is(err, service.ErrFileHashMismatch) {
		response.Error(c, http.StatusBadRequest, "FILE_HASH_MISMATCH", "sha256 mismatch")
		return
	}
	if errors.Is(err, service.ErrFileTypeNotAllowed) {
		response.Error(c, http.StatusBadRequest, "FILE_TYPE_NOT_ALLOWED", "file content does not match declared type")
		return
	}
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to complete upload")
		return
	}
	response.OK(c, out)
}

func (h *FileHandler) GetUpload(c *gin.Context) {
	uid := c.GetString("userID")
	if uid == "" {
		response.Error(c, http.StatusUnauthorized, "AUTH_SESSION_EXPIRED", "missing user")
		return
	}
	uploadID := c.Param("uploadId")
	dto, err := h.svc.GetUploadSession(c.Request.Context(), uid, uploadID)
	if errors.Is(err, service.ErrUploadNotFound) {
		response.Error(c, http.StatusNotFound, "FILE_NOT_FOUND", "upload not found")
		return
	}
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to load upload")
		return
	}
	response.OK(c, dto)
}

func (h *FileHandler) GetFile(c *gin.Context) {
	pr, ok := middleware.GetPrincipal(c)
	if !ok || pr == nil {
		response.Error(c, http.StatusUnauthorized, "AUTH_SESSION_EXPIRED", "missing principal")
		return
	}
	uid := c.GetString("userID")
	id := c.Param("fileId")
	dto, err := h.svc.GetFile(c.Request.Context(), pr, uid, id)
	if errors.Is(err, service.ErrFileNotFound) {
		response.Error(c, http.StatusNotFound, "FILE_NOT_FOUND", "file not found")
		return
	}
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to load file")
		return
	}
	response.OK(c, dto)
}

func (h *FileHandler) DownloadFile(c *gin.Context) {
	pr, ok := middleware.GetPrincipal(c)
	if !ok || pr == nil {
		response.Error(c, http.StatusUnauthorized, "AUTH_SESSION_EXPIRED", "missing principal")
		return
	}
	uid := c.GetString("userID")
	id := c.Param("fileId")
	fo, err := h.svc.GetFileObject(c.Request.Context(), pr, uid, id)
	if errors.Is(err, service.ErrFileNotFound) {
		response.Error(c, http.StatusNotFound, "FILE_NOT_FOUND", "file not found")
		return
	}
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to load file")
		return
	}
	path := h.svc.ResolvedObjectPath(fo)
	filename := id + ".bin"
	if fo.MimeType != nil && *fo.MimeType == "application/pdf" {
		filename = id + ".pdf"
	}
	c.FileAttachment(path, filename)
}

type linkFileBody struct {
	RefType string `json:"refType" binding:"required"`
	RefID   string `json:"refId" binding:"required"`
}

func (h *FileHandler) LinkFile(c *gin.Context) {
	pr, ok := middleware.GetPrincipal(c)
	if !ok || pr == nil {
		response.Error(c, http.StatusUnauthorized, "AUTH_SESSION_EXPIRED", "missing principal")
		return
	}
	uid := c.GetString("userID")
	fileID := c.Param("fileId")
	var body linkFileBody
	if err := c.ShouldBindJSON(&body); err != nil {
		response.Error(c, http.StatusBadRequest, "VALIDATION_ERROR", "invalid request body")
		return
	}
	err := h.svc.LinkFile(c.Request.Context(), uid, pr, fileID, service.LinkFileInput{
		RefType: body.RefType,
		RefID:   body.RefID,
	}, auditRequestMeta(c))
	if errors.Is(err, service.ErrFileNotFound) {
		response.Error(c, http.StatusNotFound, "FILE_NOT_FOUND", "file or target not found")
		return
	}
	if errors.Is(err, service.ErrForbiddenScope) {
		response.Error(c, http.StatusForbidden, "FORBIDDEN_SCOPE", "no access to target scope")
		return
	}
	if err != nil {
		response.Error(c, http.StatusBadRequest, "VALIDATION_ERROR", err.Error())
		return
	}
	c.Status(http.StatusNoContent)
}

func (h *FileHandler) ListFiles(c *gin.Context) {
	pr, ok := middleware.GetPrincipal(c)
	if !ok || pr == nil {
		response.Error(c, http.StatusUnauthorized, "AUTH_SESSION_EXPIRED", "missing principal")
		return
	}
	uid := c.GetString("userID")
	page, pageSize, offset := ParsePagination(c)
	items, total, err := h.svc.ListFiles(c.Request.Context(), pr, uid, offset, pageSize)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to list files")
		return
	}
	response.OK(c, gin.H{
		"items":    items,
		"total":    total,
		"page":     page,
		"pageSize": pageSize,
	})
}
