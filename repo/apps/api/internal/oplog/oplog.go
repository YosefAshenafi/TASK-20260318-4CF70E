package oplog

import (
	"encoding/json"
	"log"
	"os"
	"time"
)

var logger = log.New(os.Stdout, "", 0)

type Entry struct {
	Time      string `json:"time"`
	Level     string `json:"level"`
	Event     string `json:"event"`
	RequestID string `json:"requestId,omitempty"`
	UserID    string `json:"userId,omitempty"`
	IP        string `json:"ip,omitempty"`
	Detail    string `json:"detail,omitempty"`
}

func emit(level, event, requestID, userID, ip, detail string) {
	e := Entry{
		Time:      time.Now().UTC().Format(time.RFC3339),
		Level:     level,
		Event:     event,
		RequestID: requestID,
		UserID:    userID,
		IP:        ip,
		Detail:    detail,
	}
	b, _ := json.Marshal(e)
	logger.Println(string(b))
}

func AuthFailure(requestID, ip, username, reason string) {
	emit("WARN", "auth.failure", requestID, "", ip, reason+" user="+username)
}

func AuthSuccess(requestID, userID, ip string) {
	emit("INFO", "auth.success", requestID, userID, ip, "")
}

func PermissionDenied(requestID, userID, ip, permission string) {
	emit("WARN", "authz.permission_denied", requestID, userID, ip, "missing="+permission)
}

func ScopeViolation(requestID, userID, ip, detail string) {
	emit("WARN", "authz.scope_violation", requestID, userID, ip, detail)
}

func SessionInvalid(requestID, ip, reason string) {
	emit("WARN", "auth.session_invalid", requestID, "", ip, reason)
}

func AuditWrite(requestID, userID, module, operation, targetType, targetID string) {
	emit("INFO", "audit.write", requestID, userID, "",
		"module="+module+" op="+operation+" target="+targetType+"/"+targetID)
}

func PIIAccess(requestID, userID, candidateID string) {
	emit("INFO", "pii.access", requestID, userID, "", "candidate="+candidateID)
}

func EncryptionError(requestID, detail string) {
	emit("ERROR", "crypto.error", requestID, "", "", detail)
}
