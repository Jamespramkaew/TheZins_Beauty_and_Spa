// Package integration เก็บ integration test ที่ต่อ PostgreSQL จริง
//
// ไฟล์ทดสอบมี build tag "integration" จึงไม่ถูกรันโดย `go test ./...` ปกติ
// รันด้วย: go test -tags integration -count=1 ./test/integration/...  (หรือ make test-integration)
package integration
