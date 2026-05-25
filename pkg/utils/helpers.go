package utils

import (
    "crypto/rand"
    "encoding/hex"
    "strings"
)

func GenerateTrackingCode() string {
    bytes := make([]byte, 4)
    rand.Read(bytes)
    return strings.ToUpper(hex.EncodeToString(bytes))
}

func IsValidStatus(status string) bool {
    valid := []string{
        "pending_governor", "investigation_assigned", "investigation_done",
        "governor_processing", "process_report_submitted", "process_report_verified",
        "completion_report_submitted", "completed",
    }
    for _, s := range valid {
        if s == status {
            return true
        }
    }
    return false
}

func IsValidRole(role string) bool {
    valid := []string{"user", "investigator", "governor", "admin"}
    for _, r := range valid {
        if r == role {
            return true
        }
    }
    return false
}

func NullIfEmpty(s string) interface{} {
    if s == "" {
        return nil
    }
    return s
}