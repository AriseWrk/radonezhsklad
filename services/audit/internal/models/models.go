package models

import (
"time"

"github.com/google/uuid"
)

type AuditLog struct {
ID          uuid.UUID  `json:"id"`
UserID      *uuid.UUID `json:"user_id,omitempty"`
UserEmail   *string    `json:"user_email,omitempty"`
Method      string     `json:"method"`
Path        string     `json:"path"`
Resource    *string    `json:"resource,omitempty"`
ResourceID  *string    `json:"resource_id,omitempty"`
Status      int        `json:"status"`
RequestBody *string    `json:"request_body,omitempty"`
ClientIP    *string    `json:"client_ip,omitempty"`
RequestID   *string    `json:"request_id,omitempty"`
CreatedAt   time.Time  `json:"created_at"`
}