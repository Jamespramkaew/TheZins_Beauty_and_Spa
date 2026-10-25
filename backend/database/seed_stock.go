package database

import (
	"fmt"
	"log"
	"time"

	"backend/internal/models"

	"gorm.io/gorm"
)

// ชื่อสินค้าตัวอย่าง (ใช้ร่วมกันระหว่าง SeedStocks กับ SeedStockInventory)
const (
	stockCream     = "ครีมบำรุงหน้า"
	stockSerum     = "เซรั่มวิตามินซี"
	stockSunscreen = "ครีมกันแดด"
	stockGel       = "เจลนวดหน้า (ใช้ในร้าน)" // uses_per_unit = 20
	stockMask      = "มาส์กหน้า (ใช้ในร้าน)"  // uses_per_unit = 10
)

// receiveDaysAgo = รับของเข้าเมื่อกี่วันก่อน (ต้องเก่ากว่าการเปิดกระปุกทุกใบ)
const receiveDaysAgo = 30

// seedReceiveNote = note ของแถวรับของเข้าตั้งต้น
const seedReceiveNote = "รับของเข้าตั้งต้น (seed)"

// seedReceive = รับของเข้า 1 รายการ → stock_logs (receive, +qty)
type seedReceive struct {
	branch string // รหัสสาขา เช่น "S1"
	stock  string // ชื่อสินค้า
	qty    int
}

// seedContainer = กระปุกที่เปิดใช้ 1 ใบ → stock_containers + stock_logs (open, -1)
type seedContainer struct {
	branch    string
	stock     string
	openedAgo int    // เปิดเมื่อกี่วันก่อน
	closedAgo int    // ปิดเมื่อกี่วันก่อน, 0 = ยังเปิดอยู่
	reason    string // finished / damaged / expired (เฉพาะที่ปิดแล้ว)
	usesCount int    // จำนวนครั้งที่ใช้ไปตอนปิด (เฉพาะที่ปิดแล้ว)
}

// ยอดคงเหลือที่ได้ (receive - open) ตามข้อมูลด้านล่าง:
//
//	ของชิ้นเดียว       S1 / S2 / S3
//	  ครีมบำรุงหน้า    10 / 10 / 10
//	  เซรั่มวิตามินซี     6 /  6 /  6
//	  ครีมกันแดด       12 / 12 /  4  (S3 ต่ำกว่า min_qty = 5)
//	ของเปิดเป็นกระปุก
//	  เจลนวดหน้า        5-3=2 / 4-2=2 / 2-1=1  (S3 ต่ำกว่า min_qty = 2)
//	  มาส์กหน้า          - / 3-1=2 / 3-1=2    (S1 ไม่มีสินค้านี้เลย)
var seedReceives = []seedReceive{
	{"S1", stockCream, 10}, {"S2", stockCream, 10}, {"S3", stockCream, 10},
	{"S1", stockSerum, 6}, {"S2", stockSerum, 6}, {"S3", stockSerum, 6},
	{"S1", stockSunscreen, 12}, {"S2", stockSunscreen, 12}, {"S3", stockSunscreen, 4},
	{"S1", stockGel, 5}, {"S2", stockGel, 4}, {"S3", stockGel, 2},
	{"S2", stockMask, 3}, {"S3", stockMask, 3},
}

// กระปุกตัวอย่าง ครอบคลุม: ปิด finished / damaged / expired, เปิดอยู่, เปิดซ้อนกัน (S1 เจล 2 ใบ),
// สาขาที่ไม่มีกระปุกเปิดอยู่ (S3 เจล) และสาขาที่ไม่มีสินค้านั้น (S1 มาส์ก)
var seedContainers = []seedContainer{
	{branch: "S1", stock: stockGel, openedAgo: 20, closedAgo: 8, reason: "finished", usesCount: 20},
	{branch: "S1", stock: stockGel, openedAgo: 5},
	{branch: "S1", stock: stockGel, openedAgo: 1}, // เปิดซ้อน: การใช้นับให้ใบนี้ (เปิดล่าสุด)

	{branch: "S2", stock: stockGel, openedAgo: 18, closedAgo: 6, reason: "finished", usesCount: 20},
	{branch: "S2", stock: stockGel, openedAgo: 4},
	{branch: "S2", stock: stockMask, openedAgo: 25, closedAgo: 7, reason: "expired", usesCount: 4},

	{branch: "S3", stock: stockGel, openedAgo: 10, closedAgo: 9, reason: "damaged", usesCount: 7},
	{branch: "S3", stock: stockMask, openedAgo: 2},
}

// SeedStockInventory สร้างสต็อกตั้งต้นให้สอดคล้องกันทั้งระบบ ในทรานแซกชันเดียว:
//   - รับของเข้า → stock_logs (receive)
//   - เปิดกระปุก 1 ใบ = แถวใน stock_containers + stock_logs (open, -1) คู่กันเสมอ
//   - ปิดเพราะ damaged / expired ไม่สร้าง stock_logs เพิ่ม (ตัดไปแล้วตอนเปิด)
//
// ทุกแถวมี employee_id ของพนักงานในสาขาเดียวกับแถวนั้น: รับของเข้า = manager, เปิดกระปุก = TP
// รับของเข้ามี note อธิบายที่มา เปิดกระปุกไม่มี note
//
// ถ้าตาราง stock_logs หรือ stock_containers มีข้อมูลอยู่แล้วจะข้ามทั้งหมด (กันข้อมูลซ้ำ)
// ต้องการสร้างใหม่ให้รัน seed ด้วย -reset ต้องรันหลัง SeedBranches, SeedEmployees และ SeedStocks
//
// หมายเหตุ: uses_count ของกระปุกที่ปิดแล้วเป็นค่าสมมติ เพราะระบบนี้ยังไม่ seed service_usages
// (ปกติค่านี้คำนวณจาก service_usages ในช่วง opened_at ถึง closed_at)
func SeedStockInventory(db *gorm.DB) error {
	var logCount, containerCount int64
	if err := db.Model(&models.StockLog{}).Count(&logCount).Error; err != nil {
		return err
	}
	if err := db.Model(&models.StockContainer{}).Count(&containerCount).Error; err != nil {
		return err
	}
	if logCount > 0 || containerCount > 0 {
		log.Println("Stock inventory already exists, skipping (run seed with -reset to reseed)")
		return nil
	}

	branchByCode, err := loadBranchesByCode(db)
	if err != nil {
		return err
	}
	stockByName, err := loadStocksByName(db)
	if err != nil {
		return err
	}
	employeeByBranchRole, err := loadEmployeesByBranchRole(db)
	if err != nil {
		return err
	}

	now := time.Now()

	return db.Transaction(func(tx *gorm.DB) error {
		var logs []models.StockLog
		var containers []models.StockContainer

		for _, r := range seedReceives {
			branch, ok := branchByCode[r.branch]
			if !ok {
				return fmt.Errorf("seed receive: branch %q not found", r.branch)
			}
			stock, ok := stockByName[r.stock]
			if !ok {
				return fmt.Errorf("seed receive: stock %q not found", r.stock)
			}
			if r.qty <= 0 {
				return fmt.Errorf("seed receive: qty must be > 0 (%s / %s)", r.branch, r.stock)
			}

			emp, err := seedEmployeeFor(employeeByBranchRole, branch.ID, "manager")
			if err != nil {
				return fmt.Errorf("seed receive: %w", err)
			}

			note := seedReceiveNote
			logs = append(logs, models.StockLog{
				StockID:    stock.ID,
				BranchID:   branch.ID,
				EmployeeID: emp.ID,
				QtyChange:  r.qty,
				LogType:    "receive",
				Note:       &note,
				CreatedAt:  now.AddDate(0, 0, -receiveDaysAgo),
			})
		}

		for _, s := range seedContainers {
			branch, ok := branchByCode[s.branch]
			if !ok {
				return fmt.Errorf("seed container: branch %q not found", s.branch)
			}
			stock, ok := stockByName[s.stock]
			if !ok {
				return fmt.Errorf("seed container: stock %q not found", s.stock)
			}
			if err := validateSeedContainer(s, stock); err != nil {
				return err
			}

			emp, err := seedEmployeeFor(employeeByBranchRole, branch.ID, "TP")
			if err != nil {
				return fmt.Errorf("seed container: %w", err)
			}

			openedAt := now.AddDate(0, 0, -s.openedAgo)
			container := models.StockContainer{
				StockID:  stock.ID,
				BranchID: branch.ID,
				OpenedAt: openedAt,
			}
			if s.closedAgo > 0 {
				closedAt := now.AddDate(0, 0, -s.closedAgo)
				reason := s.reason
				uses := s.usesCount
				container.ClosedAt, container.CloseReason, container.UsesCount = &closedAt, &reason, &uses
			}
			containers = append(containers, container)

			// เปิดกระปุก = ตัดสต็อกของที่ยังไม่เปิดออก 1 ชิ้น
			logs = append(logs, models.StockLog{
				StockID:    stock.ID,
				BranchID:   branch.ID,
				EmployeeID: emp.ID,
				QtyChange:  -1,
				LogType:    "open",
				CreatedAt:  openedAt,
			})
		}

		if err := tx.Create(&logs).Error; err != nil {
			return err
		}
		if err := tx.Create(&containers).Error; err != nil {
			return err
		}

		// ตรวจความสอดคล้องก่อน commit ถ้าไม่ผ่านจะ rollback ทั้งหมด
		return verifyStockInventory(tx)
	})
}

func loadBranchesByCode(db *gorm.DB) (map[string]models.Branch, error) {
	var branches []models.Branch
	if err := db.Find(&branches).Error; err != nil {
		return nil, err
	}
	m := make(map[string]models.Branch, len(branches))
	for _, b := range branches {
		m[b.Code] = b
	}
	return m, nil
}

// loadEmployeesByBranchRole คืนพนักงานที่ active แยกตาม "branchID|role" (ถ้ามีหลายคนใช้คนแรกตามชื่อ)
func loadEmployeesByBranchRole(db *gorm.DB) (map[string]models.Employee, error) {
	var employees []models.Employee
	if err := db.Where("is_active = ?", true).Order("name DESC").Find(&employees).Error; err != nil {
		return nil, err
	}
	m := make(map[string]models.Employee, len(employees))
	for _, e := range employees {
		m[e.BranchID+"|"+e.Role] = e // เรียง DESC เพื่อให้คนแรกตามชื่อเขียนทับเป็นคนสุดท้าย
	}
	return m, nil
}

func seedEmployeeFor(m map[string]models.Employee, branchID, role string) (models.Employee, error) {
	e, ok := m[branchID+"|"+role]
	if !ok {
		return models.Employee{}, fmt.Errorf("no active %q employee in branch %s (run SeedEmployees first)", role, branchID)
	}
	return e, nil
}

func loadStocksByName(db *gorm.DB) (map[string]models.Stock, error) {
	var stocks []models.Stock
	if err := db.Find(&stocks).Error; err != nil {
		return nil, err
	}
	m := make(map[string]models.Stock, len(stocks))
	for _, s := range stocks {
		m[s.Name] = s
	}
	return m, nil
}

// validateSeedContainer ตรวจข้อมูลตัวอย่างก่อนบันทึก ให้ตรงกติกาของ StockContainer
func validateSeedContainer(s seedContainer, stock models.Stock) error {
	where := fmt.Sprintf("seed container (%s / %s)", s.branch, s.stock)

	if stock.UsesPerUnit <= 1 {
		return fmt.Errorf("%s: only stocks with uses_per_unit > 1 can have containers", where)
	}
	if s.openedAgo <= 0 || s.openedAgo >= receiveDaysAgo {
		return fmt.Errorf("%s: openedAgo must be between 1 and %d (must be after receiving)", where, receiveDaysAgo-1)
	}

	if s.closedAgo == 0 {
		// ยังเปิดอยู่: ต้องไม่มี reason / usesCount
		if s.reason != "" || s.usesCount != 0 {
			return fmt.Errorf("%s: an open container must not have reason or usesCount", where)
		}
		return nil
	}

	if s.closedAgo < 0 || s.closedAgo >= s.openedAgo {
		return fmt.Errorf("%s: closedAgo must be after opened and not in the future", where)
	}
	switch s.reason {
	case "finished", "damaged", "expired":
	default:
		return fmt.Errorf("%s: invalid close reason %q", where, s.reason)
	}
	if s.usesCount < 1 || s.usesCount > stock.UsesPerUnit {
		return fmt.Errorf("%s: usesCount must be between 1 and %d", where, stock.UsesPerUnit)
	}
	return nil
}

// verifyStockInventory ตรวจข้อมูลที่ seed จริงใน DB (เรียกภายในทรานแซกชัน)
func verifyStockInventory(tx *gorm.DB) error {
	checks := []struct {
		name  string
		query string // ต้องคืนจำนวนแถวที่ "ผิด" ถ้าไม่ใช่ 0 แปลว่าไม่ผ่าน
	}{
		{
			"stock balance must never be negative",
			`SELECT COUNT(*) FROM (
				SELECT 1 FROM stock_logs GROUP BY branch_id, stock_id HAVING SUM(qty_change) < 0
			) t`,
		},
		{
			"every container needs exactly one matching 'open' log per branch and stock",
			`SELECT COUNT(*) FROM
				(SELECT branch_id, stock_id, COUNT(*) AS n FROM stock_logs WHERE log_type = 'open' GROUP BY branch_id, stock_id) l
				FULL JOIN
				(SELECT branch_id, stock_id, COUNT(*) AS n FROM stock_containers GROUP BY branch_id, stock_id) c
				USING (branch_id, stock_id)
			WHERE COALESCE(l.n, 0) <> COALESCE(c.n, 0)`,
		},
		{
			"containers must belong to stocks with uses_per_unit > 1",
			`SELECT COUNT(*) FROM stock_containers c
				JOIN stocks s ON s.id = c.stock_id
			WHERE s.uses_per_unit <= 1`,
		},
		{
			"'open' log must be -1 and 'receive' log must be positive",
			`SELECT COUNT(*) FROM stock_logs
			WHERE (log_type = 'open' AND qty_change <> -1) OR (log_type = 'receive' AND qty_change <= 0)`,
		},
		{
			"every log must be made by an active employee of the same branch (transfer: the sender)",
			`SELECT COUNT(*) FROM stock_logs l
				LEFT JOIN employees e ON e.id = l.employee_id
			WHERE e.id IS NULL OR e.is_active = false
			   OR (e.branch_id <> l.branch_id AND l.log_type <> 'transfer')`,
		},
		{
			"adjust and count_diff logs need a note",
			`SELECT COUNT(*) FROM stock_logs
			WHERE log_type IN ('adjust', 'count_diff') AND (note IS NULL OR btrim(note) = '')`,
		},
	}

	for _, c := range checks {
		var bad int64
		if err := tx.Raw(c.query).Scan(&bad).Error; err != nil {
			return fmt.Errorf("verify %q: %w", c.name, err)
		}
		if bad != 0 {
			return fmt.Errorf("seed verification failed: %s (%d rows)", c.name, bad)
		}
	}
	return nil
}
