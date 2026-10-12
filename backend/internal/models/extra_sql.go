package models

import _ "embed"

// ExtraSQL = SQL เสริมที่ AutoMigrate ทำไม่ได้ (ดู extra.sql)
// รันหลัง AutoMigrate ทุกครั้ง: db.Exec(models.ExtraSQL) รันซ้ำได้ปลอดภัย
//
//go:embed extra.sql
var ExtraSQL string
