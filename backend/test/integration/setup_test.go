//go:build integration

package integration

import (
	"errors"
	"fmt"
	"log"
	"os"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"backend/database"
	"backend/internal/models"

	"github.com/joho/godotenv"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

// db = connection pool ที่ใช้ร่วมกันทั้ง package (เปิดครั้งเดียวใน TestMain)
var db *gorm.DB

// ความปลอดภัยของ DB:
//   - ทุก test ทำงานใน transaction เดียวแล้ว ROLLBACK ตอนจบ จึงไม่ทิ้งข้อมูลใน DB เลย
//     (แม้ test พังกลางทางหรือโปรเซสถูกฆ่า connection หลุด Postgres ก็ rollback ให้เอง)
//   - ไม่มีการ DROP / TRUNCATE / AutoMigrate ใดๆ
//   - ปฏิเสธการรันถ้า DB ไม่ใช่เครื่องตัวเอง หรือ ENVIRONMENT=production
func TestMain(m *testing.M) {
	// ไฟล์ .env ของ backend อยู่สองชั้นขึ้นไปจากโฟลเดอร์นี้ (ไม่เจอก็ใช้ env ของระบบ)
	_ = godotenv.Load("../../.env")

	if strings.EqualFold(os.Getenv("ENVIRONMENT"), "production") {
		log.Fatal("integration test: ปฏิเสธการรันเมื่อ ENVIRONMENT=production")
	}

	host := getenv("DB_HOST", "localhost")
	if host != "localhost" && host != "127.0.0.1" && os.Getenv("INTEGRATION_ALLOW_REMOTE_DB") != "1" {
		log.Fatalf("integration test: DB_HOST=%q ไม่ใช่เครื่องตัวเอง ถ้าตั้งใจจริงให้ตั้ง INTEGRATION_ALLOW_REMOTE_DB=1", host)
	}

	// INTEGRATION_DB_NAME ใช้ชี้ไป DB แยกสำหรับ test (make test-integration-isolated)
	// ไม่ตั้ง = ใช้ DB_NAME จาก .env (DB dev) ซึ่งปลอดภัยเพราะทุก test rollback
	dbName := getenv("INTEGRATION_DB_NAME", getenv("DB_NAME", "the_zins_db"))

	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		host,
		getenv("DB_PORT", "5432"),
		getenv("DB_USER", "postgres"),
		getenv("DB_PASSWORD", ""),
		dbName,
		getenv("DB_SSLMODE", "disable"),
	)

	var err error
	db, err = database.Connect(dsn) // เปิด TranslateError เหมือนระบบจริง (ErrDuplicatedKey ฯลฯ)
	if err != nil {
		log.Fatalf("integration test: เชื่อมต่อ DB %q ไม่ได้ (container รันอยู่ไหม? ลอง make docker-up): %v", dbName, err)
	}

	// DB ที่ลงท้าย _test เป็นของ test โดยเฉพาะ สร้าง schema ให้เองได้ (AutoMigrate + extra.sql ไม่ลบอะไร)
	// DB อื่น (เช่น dev) ไม่แตะ schema เด็ดขาด แค่ตรวจว่า migrate ไว้แล้ว
	if strings.HasSuffix(dbName, "_test") {
		if err := database.RunMigration(db); err != nil {
			log.Fatalf("integration test: migrate DB %q ไม่สำเร็จ: %v", dbName, err)
		}
	}

	if err := requireSchema(db); err != nil {
		log.Fatalf("integration test: %v", err)
	}

	code := m.Run()
	_ = database.Close(db)
	os.Exit(code)
}

func getenv(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

// requireSchema เช็คว่า DB ถูก migrate แล้ว (ตาราง + unique index ที่ test พึ่งพา) โดยไม่แก้ schema เอง
func requireSchema(db *gorm.DB) error {
	for _, table := range []string{"stocks", "stock_containers", "stock_logs", "order_items", "course_items", "branches"} {
		var name *string
		if err := db.Raw("SELECT to_regclass(?)::text", table).Scan(&name).Error; err != nil {
			return err
		}
		if name == nil {
			return fmt.Errorf("ไม่พบตาราง %s: รัน migration ก่อน (make migrate)", table)
		}
	}

	var n int64
	err := db.Raw("SELECT count(*) FROM pg_indexes WHERE indexname = 'idx_stocks_name_active'").Scan(&n).Error
	if err != nil {
		return err
	}
	if n == 0 {
		return errors.New("ไม่พบ index idx_stocks_name_active: รัน migration ให้ครบ (extra.sql) ก่อน")
	}
	return nil
}

// newTx เริ่ม transaction ให้ test นี้ และ rollback เมื่อจบ ทุกอย่างที่สร้างจะหายไป
func newTx(t *testing.T) *gorm.DB {
	t.Helper()
	tx := db.Begin()
	require.NoError(t, tx.Error)
	t.Cleanup(func() { tx.Rollback() })
	return tx
}

// expectDBError รันคำสั่งที่ตั้งใจให้ DB ปฏิเสธ (unique/FK)
// Postgres จะทำให้ทั้ง transaction ใช้ต่อไม่ได้หลัง error จึงต้องครอบด้วย savepoint แล้วย้อนกลับ
func expectDBError(t *testing.T, tx *gorm.DB, fn func() error) error {
	t.Helper()
	require.NoError(t, tx.SavePoint("sp").Error)
	err := fn()
	if err != nil {
		require.NoError(t, tx.RollbackTo("sp").Error)
	}
	return err
}

// ---------- ตัวสร้างข้อมูลทดสอบ ----------

var seq uint64

// tag คืนเลข 8 หลักที่ไม่ซ้ำกันในการรันหนึ่งๆ ใช้ทำชื่อ/รหัสที่ต้อง unique
func tag() string {
	n := atomic.AddUint64(&seq, 1)
	return fmt.Sprintf("%06d%02d", (time.Now().UnixNano()/1000)%1_000_000, n%100)
}

func mkBranch(t *testing.T, tx *gorm.DB) models.Branch {
	t.Helper()
	b := models.Branch{Code: "T" + tag(), Name: "itest branch", IsActive: true}
	require.NoError(t, tx.Create(&b).Error)
	return b
}

func mkStock(t *testing.T, tx *gorm.DB, name string) models.Stock {
	t.Helper()
	s := models.Stock{
		Name:        name,
		Price:       decimal.NewFromInt(100),
		IsSellable:  true,
		Category:    "normal",
		UsesPerUnit: 3,
		IsActive:    true,
	}
	require.NoError(t, tx.Create(&s).Error)
	return s
}

func uniqueName(prefix string) string { return prefix + "-" + tag() }

// mkOrderItem สร้างบิล 1 ใบ (พร้อมลูกค้า/สาขา) ที่มีรายการขายสินค้านี้ 1 บรรทัด
func mkOrderItem(t *testing.T, tx *gorm.DB, stockID string) models.OrderItem {
	t.Helper()
	branch := mkBranch(t, tx)
	customer := models.Customer{Fname: "itest", Tier: "general", HnNo: "T" + tag()}
	require.NoError(t, tx.Create(&customer).Error)
	order := models.Order{
		OrderNo:    "T" + tag(),
		CustomerID: customer.ID,
		BranchID:   branch.ID,
		OrderDate:  time.Now(),
	}
	require.NoError(t, tx.Create(&order).Error)

	item := models.OrderItem{
		OrderID:   order.ID,
		LineNo:    1,
		ItemType:  "stock",
		StockID:   &stockID,
		Qty:       1,
		UnitPrice: decimal.NewFromInt(100),
	}
	require.NoError(t, tx.Create(&item).Error)
	return item
}

func mkCourseItem(t *testing.T, tx *gorm.DB, stockID string) models.CourseItem {
	t.Helper()
	course := models.Course{Name: uniqueName("itest-course"), Price: decimal.NewFromInt(100), MaxSession: 1, IsActive: true}
	require.NoError(t, tx.Create(&course).Error)
	ci := models.CourseItem{CourseID: course.ID, StockID: stockID, UseQty: 1}
	require.NoError(t, tx.Create(&ci).Error)
	return ci
}

func mkEmployee(t *testing.T, tx *gorm.DB, branchID, role string) models.Employee {
	t.Helper()
	e := models.Employee{Name: uniqueName("itest-emp"), Role: role, BranchID: branchID, IsActive: true}
	require.NoError(t, tx.Create(&e).Error)
	return e
}

// mkStockLog สร้าง log รับของเข้า +5 โดยพนักงานใหม่ของสาขานั้น (employee_id บังคับ NOT NULL)
func mkStockLog(t *testing.T, tx *gorm.DB, stockID, branchID string) models.StockLog {
	t.Helper()
	emp := mkEmployee(t, tx, branchID, "manager")
	return mkStockLogBy(t, tx, stockID, branchID, emp.ID, "receive", 5)
}

// mkStockLogBy สร้าง log ตามที่ระบุ (ใช้ในเทสที่ต้องคุมพนักงาน/ชนิด/จำนวนเอง)
func mkStockLogBy(t *testing.T, tx *gorm.DB, stockID, branchID, employeeID, logType string, qty int) models.StockLog {
	t.Helper()
	l := models.StockLog{StockID: stockID, BranchID: branchID, EmployeeID: employeeID, QtyChange: qty, LogType: logType}
	require.NoError(t, tx.Create(&l).Error)
	return l
}

// mkContainer: closed = true จะใส่ closed_at / close_reason / uses_count ให้ผ่าน CHECK constraint
func mkContainer(t *testing.T, tx *gorm.DB, stockID, branchID string, openedAt time.Time, closed bool) models.StockContainer {
	t.Helper()
	c := models.StockContainer{StockID: stockID, BranchID: branchID, OpenedAt: openedAt}
	if closed {
		closedAt := openedAt.Add(time.Minute)
		reason := "finished"
		uses := 3
		c.ClosedAt, c.CloseReason, c.UsesCount = &closedAt, &reason, &uses
	}
	require.NoError(t, tx.Create(&c).Error)
	return c
}
