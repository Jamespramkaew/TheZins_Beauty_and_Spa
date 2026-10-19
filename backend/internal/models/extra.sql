-- SQL เสริมที่ AutoMigrate ทำไม่ได้
-- รันหลัง AutoMigrate ทุกครั้งที่เปิด server ผ่าน db.Exec(models.ExtraSQL)
-- ทุกคำสั่งเขียนให้รันซ้ำได้: ถ้ามีอยู่แล้วจะข้ามไป ไม่ error
-- ถ้าแก้เงื่อนไขของ constraint ที่มีอยู่แล้ว ให้ตั้งชื่อใหม่ หรือ DROP ตัวเก่าในฐานข้อมูลก่อน

-- จำเป็นสำหรับ EXCLUDE ด้านล่าง (ให้ GiST index ใช้ room_id WITH = ได้)
CREATE EXTENSION IF NOT EXISTS btree_gist;

-- order_items: บังคับคอลัมน์ตามประเภทรายการ
--   course → ต้องมี course_id
--   stock  → ต้องมี stock_id
--   credit → ต้องมี credit_type และ credit_amount
DO $$
BEGIN
  IF NOT EXISTS (SELECT 1 FROM pg_constraint WHERE conname = 'chk_order_items_type') THEN
    ALTER TABLE order_items ADD CONSTRAINT chk_order_items_type CHECK (
      (item_type = 'course' AND course_id IS NOT NULL) OR
      (item_type = 'stock'  AND stock_id  IS NOT NULL) OR
      (item_type = 'credit' AND credit_type IS NOT NULL AND credit_amount IS NOT NULL)
    );
  END IF;
END $$;

-- appointments: ห้องเดียวกันห้ามจองเวลาทับกัน (นับเฉพาะคิวที่ booked / done)
-- ช่วงเวลาเป็นแบบ [start_at, end_at) จึงจองต่อกันพอดีได้ เช่น 10:00-11:00 กับ 11:00-12:00
DO $$
BEGIN
  IF NOT EXISTS (SELECT 1 FROM pg_constraint WHERE conname = 'excl_appointments_room_time') THEN
    ALTER TABLE appointments ADD CONSTRAINT excl_appointments_room_time
      EXCLUDE USING gist (room_id WITH =, tstzrange(start_at, end_at) WITH &&)
      WHERE (status IN ('booked', 'done'));
  END IF;
END $$;

-- stock_logs: ให้คำนวณยอดคงเหลือ SUM(qty_change) จาก index ได้เลยโดยไม่ต้องอ่านตาราง
-- ลำดับ: stock_id ก่อน branch_id เพราะ stock_id มี selectivity สูงกว่า (50-100 สินค้า vs 3 สาขา)
-- INCLUDE (qty_change) ทำให้ PostgreSQL ทำ index-only scan ได้ (ไม่ต้องไปหาตารางหลัก)
CREATE INDEX IF NOT EXISTS idx_stock_logs_balance ON stock_logs (stock_id, branch_id) INCLUDE (qty_change);

-- Index เสริมสำหรับ query ที่ filter เฉพาะ stock_id (admin ดูรวมทุกสาขา)
CREATE INDEX IF NOT EXISTS idx_stock_logs_stock_qty ON stock_logs (stock_id) INCLUDE (qty_change);
