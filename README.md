# TheZins Beauty & Spa

เว็บไซต์ร้านสปา แยกเป็น frontend และ backend อยู่ใน repo เดียวกัน (monorepo)

## โครงสร้างโปรเจกต์

```
TheZins_Beauty_and_Spa/
├── frontend/   ← React + Vite
└── backend/    ← Backend API
```

## Frontend

### Tech stack

React, Vite, React Router, Axios, React Hook Form + Zod, Tailwind CSS, Recharts, Swiper, React Icons, Day.js

### เริ่มต้นใช้งาน

```bash
cd frontend
npm install
cp .env.example .env   # แล้วแก้ค่าให้ตรงกับเครื่องตัวเอง
npm run dev            # เปิด http://localhost:5173
```

### Environment variables

| ตัวแปร | คำอธิบาย | ตัวอย่าง |
|---|---|---|
| `VITE_APP_NAME` | ชื่อแอปที่แสดงบนหน้าเว็บ | `TheZins Beauty & Spa` |
| `VITE_API_BASE_URL` | URL ของ Backend API | `http://localhost:3000/api` |
| `VITE_API_TIMEOUT` | เวลารอ API สูงสุด (มิลลิวินาที) | `10000` |

> ⚠️ ค่าใน `.env` ของ frontend จะถูกฝังลงในโค้ดที่ส่งให้เบราว์เซอร์ ห้ามใส่ข้อมูลลับ

### Scripts

| คำสั่ง | ทำอะไร |
|---|---|
| `npm run dev` | รัน dev server |
| `npm run build` | build สำหรับ production (ได้โฟลเดอร์ `dist/`) |
| `npm run preview` | เปิดดูผลการ build |
| `npm run lint` | ตรวจโค้ดด้วย ESLint |

## Backend

ดูวิธีติดตั้งและรันได้ที่ `backend/README.md`

> Backend ต้องเปิด CORS ให้ `http://localhost:5173` เพื่อให้ frontend เรียก API ได้ตอนพัฒนา
