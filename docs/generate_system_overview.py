#!/usr/bin/env python3.12
"""
SekolahPro System Overview - PDF Generator
Generates a comprehensive system visualization with architecture,
flow diagrams, policies, and sprint roadmap.
"""

from fpdf import FPDF
import os

# ─── Constants ────────────────────────────────────────────────────────────────

OUTPUT_PATH = os.path.join(os.path.dirname(__file__), "SekolahPro-System-Overview.pdf")

# Colors (R, G, B)
C_PRIMARY = (41, 98, 255)       # Blue
C_SECONDARY = (0, 150, 136)     # Teal
C_ACCENT = (255, 152, 0)        # Orange
C_KOPERASI = (156, 39, 176)     # Purple
C_SEKOLAH = (33, 150, 243)      # Light Blue
C_CORE = (76, 175, 80)          # Green
C_INFRA = (96, 125, 139)        # Blue Grey
C_BG_LIGHT = (248, 249, 250)    # Light grey bg
C_BG_BOX = (232, 245, 253)      # Light blue bg
C_BG_GREEN = (232, 245, 233)    # Light green bg
C_BG_PURPLE = (243, 229, 245)   # Light purple bg
C_BG_ORANGE = (255, 243, 224)   # Light orange bg
C_TEXT = (33, 33, 33)           # Dark text
C_TEXT_LIGHT = (117, 117, 117)  # Grey text
C_WHITE = (255, 255, 255)
C_BLACK = (0, 0, 0)
C_RED = (244, 67, 54)


class SystemPDF(FPDF):
    def __init__(self):
        super().__init__(orientation='L', unit='mm', format='A4')
        self.set_auto_page_break(auto=True, margin=15)

    def header(self):
        if self.page_no() > 1:
            self.set_font("Helvetica", "I", 8)
            self.set_text_color(*C_TEXT_LIGHT)
            self.cell(0, 5, "SekolahPro System Overview | Vernon Corp", align="L")
            self.cell(0, 5, f"Page {self.page_no()}", align="R", new_x="LMARGIN", new_y="NEXT")
            self.line(10, 12, 287, 12)
            self.ln(3)

    def footer(self):
        self.set_y(-10)
        self.set_font("Helvetica", "I", 7)
        self.set_text_color(*C_TEXT_LIGHT)
        self.cell(0, 5, "Generated: 2026-04-15 | Confidential - Vernon Corp Engineering", align="C")

    # ─── Drawing Helpers ──────────────────────────────────────────────────

    def draw_rounded_box(self, x, y, w, h, bg_color, border_color=None, radius=3):
        self.set_fill_color(*bg_color)
        if border_color:
            self.set_draw_color(*border_color)
            self.set_line_width(0.5)
            self.rect(x, y, w, h, style="DF")
        else:
            self.rect(x, y, w, h, style="F")

    def draw_arrow(self, x1, y1, x2, y2, color=C_TEXT_LIGHT):
        self.set_draw_color(*color)
        self.set_line_width(0.4)
        self.line(x1, y1, x2, y2)
        # arrowhead
        import math
        angle = math.atan2(y2 - y1, x2 - x1)
        arrow_len = 2.5
        self.line(x2, y2,
                  x2 - arrow_len * math.cos(angle - 0.4),
                  y2 - arrow_len * math.sin(angle - 0.4))
        self.line(x2, y2,
                  x2 - arrow_len * math.cos(angle + 0.4),
                  y2 - arrow_len * math.sin(angle + 0.4))

    def section_title(self, text, color=C_PRIMARY, y_offset=0):
        self.set_font("Helvetica", "B", 16)
        self.set_text_color(*color)
        self.cell(0, 10, text, new_x="LMARGIN", new_y="NEXT")
        self.set_draw_color(*color)
        self.set_line_width(0.8)
        self.line(10, self.get_y(), 100, self.get_y())
        self.ln(4)

    def sub_title(self, text, color=C_TEXT):
        self.set_font("Helvetica", "B", 11)
        self.set_text_color(*color)
        self.cell(0, 7, text, new_x="LMARGIN", new_y="NEXT")
        self.ln(1)

    def body_text(self, text, color=C_TEXT):
        self.set_font("Helvetica", "", 9)
        self.set_text_color(*color)
        self.multi_cell(0, 4.5, text)
        self.ln(1)

    def label_in_box(self, x, y, w, h, text, bg, text_color=C_WHITE, font_size=8, bold=True):
        self.draw_rounded_box(x, y, w, h, bg)
        self.set_xy(x, y + 0.5)
        style = "B" if bold else ""
        self.set_font("Helvetica", style, font_size)
        self.set_text_color(*text_color)
        self.cell(w, h - 1, text, align="C")


def build_pdf():
    pdf = SystemPDF()

    # ═══════════════════════════════════════════════════════════════════════
    # PAGE 1: COVER
    # ═══════════════════════════════════════════════════════════════════════
    pdf.add_page()
    pdf.set_font("Helvetica", "B", 36)
    pdf.set_text_color(*C_PRIMARY)
    pdf.ln(35)
    pdf.cell(0, 15, "SekolahPro", align="C", new_x="LMARGIN", new_y="NEXT")

    pdf.set_font("Helvetica", "", 18)
    pdf.set_text_color(*C_TEXT_LIGHT)
    pdf.cell(0, 10, "System Architecture & Flow Overview", align="C", new_x="LMARGIN", new_y="NEXT")
    pdf.ln(8)

    pdf.set_draw_color(*C_PRIMARY)
    pdf.set_line_width(1)
    pdf.line(80, pdf.get_y(), 217, pdf.get_y())
    pdf.ln(10)

    pdf.set_font("Helvetica", "", 12)
    pdf.set_text_color(*C_TEXT)
    lines = [
        "Management Sekolah + Management Koperasi Sekolah",
        "Dual-Mode: General (Umum) & Islamic (Pesantren/BMT)",
        "",
        "Tech Stack: Go (Clean Architecture + CQRS) | React 18 + Vite",
        "Vernon Denormalized Read-Cache | Multi-Tenant 4-Level Hierarchy",
        "Event-Driven (NATS JetStream) | PostgreSQL + Redis",
        "",
        "100 ADRs | 18 Core + 58 Sekolah + 24 Koperasi",
        "15 Sprint Phases | ~100+ Database Tables",
    ]
    for line in lines:
        pdf.cell(0, 7, line, align="C", new_x="LMARGIN", new_y="NEXT")

    pdf.ln(15)
    pdf.set_font("Helvetica", "I", 10)
    pdf.set_text_color(*C_TEXT_LIGHT)
    pdf.cell(0, 7, "Vernon Corp Engineering | April 2026", align="C", new_x="LMARGIN", new_y="NEXT")

    # ═══════════════════════════════════════════════════════════════════════
    # PAGE 2: HIGH-LEVEL ARCHITECTURE
    # ═══════════════════════════════════════════════════════════════════════
    pdf.add_page()
    pdf.section_title("1. High-Level System Architecture")

    # --- Draw architecture diagram ---
    # Users layer
    y_start = 30
    users = [("Admin/Kepsek", 25), ("Guru/Staff", 75), ("Orang Tua", 125), ("Siswa", 175), ("Teller/Koperasi", 225)]
    for label, x in users:
        pdf.label_in_box(x, y_start, 40, 10, label, C_INFRA)
        pdf.draw_arrow(x + 20, y_start + 10, x + 20, y_start + 17, C_TEXT_LIGHT)

    # Frontend layer
    y_fe = y_start + 18
    pdf.draw_rounded_box(15, y_fe, 262, 16, C_BG_BOX, C_SEKOLAH)
    pdf.set_xy(15, y_fe + 1)
    pdf.set_font("Helvetica", "B", 10)
    pdf.set_text_color(*C_SEKOLAH)
    pdf.cell(262, 6, "FRONTEND: React 18 + Vite + CSS Modules (Multi-App Strategy)", align="C")
    pdf.set_xy(15, y_fe + 7)
    pdf.set_font("Helvetica", "", 8)
    pdf.set_text_color(*C_TEXT)
    pdf.cell(262, 6, "App Sekolah  |  App Koperasi  |  Portal Orang Tua  |  Dashboard Kepala Sekolah  |  PWA Mobile", align="C")

    # Arrow FE -> API
    pdf.draw_arrow(148, y_fe + 16, 148, y_fe + 23, C_PRIMARY)

    # API Gateway layer
    y_api = y_fe + 24
    pdf.draw_rounded_box(15, y_api, 262, 12, C_BG_ORANGE, C_ACCENT)
    pdf.set_xy(15, y_api + 1)
    pdf.set_font("Helvetica", "B", 10)
    pdf.set_text_color(*C_ACCENT)
    pdf.cell(262, 5, "API LAYER: Chi Router + Two-Phase JWT Auth + RBAC Middleware", align="C")
    pdf.set_xy(15, y_api + 6)
    pdf.set_font("Helvetica", "", 8)
    pdf.set_text_color(*C_TEXT)
    pdf.cell(262, 5, "Multi-Tenant Scope Filter  |  Rate Limiting  |  Request Validation  |  Telemetry (OpenTelemetry)", align="C")

    # Arrow API -> Services
    pdf.draw_arrow(90, y_api + 12, 90, y_api + 19, C_SEKOLAH)
    pdf.draw_arrow(200, y_api + 12, 200, y_api + 19, C_KOPERASI)

    # Backend Services
    y_svc = y_api + 20
    # Sekolah service
    pdf.draw_rounded_box(15, y_svc, 125, 38, C_BG_BOX, C_SEKOLAH)
    pdf.set_xy(15, y_svc + 1)
    pdf.set_font("Helvetica", "B", 10)
    pdf.set_text_color(*C_SEKOLAH)
    pdf.cell(125, 6, "MANAGEMENT SEKOLAH", align="C")
    pdf.set_font("Helvetica", "", 7.5)
    pdf.set_text_color(*C_TEXT)
    items_s = [
        "Siswa & Wali  |  Kurikulum & Jadwal  |  Absensi & Nilai",
        "Rapor  |  SPP & Keuangan  |  Perpustakaan & Lab",
        "Asrama & Kantin  |  PPDB  |  Guru HR & Payroll",
        "Portal Orang Tua  |  Dapodik Integration",
    ]
    y_t = y_svc + 8
    for item in items_s:
        pdf.set_xy(17, y_t)
        pdf.cell(121, 4.5, item, align="C")
        y_t += 5

    pdf.label_in_box(48, y_svc + 30, 60, 7, "58 ADR | ~70 Tables", C_SEKOLAH, font_size=7)

    # Koperasi service
    pdf.draw_rounded_box(152, y_svc, 125, 38, C_BG_PURPLE, C_KOPERASI)
    pdf.set_xy(152, y_svc + 1)
    pdf.set_font("Helvetica", "B", 10)
    pdf.set_text_color(*C_KOPERASI)
    pdf.cell(125, 6, "MANAGEMENT KOPERASI SEKOLAH", align="C")
    pdf.set_font("Helvetica", "", 7.5)
    pdf.set_text_color(*C_TEXT)
    items_k = [
        "Nasabah & Rekening  |  Produk & Akad Syariah",
        "Simpanan (Pokok/Wajib/Tabungan/Deposito)",
        "Pinjaman & Angsuran  |  Teller & Kas  |  Jurnal COA",
        "SHU  |  E-Wallet  |  Toko POS  |  Zakat (BMT)",
    ]
    y_t = y_svc + 8
    for item in items_k:
        pdf.set_xy(154, y_t)
        pdf.cell(121, 4.5, item, align="C")
        y_t += 5

    pdf.label_in_box(185, y_svc + 30, 60, 7, "24 ADR | ~30 Tables", C_KOPERASI, font_size=7)

    # Event Bus
    y_evt = y_svc + 42
    pdf.draw_rounded_box(15, y_evt, 262, 10, C_BG_GREEN, C_CORE)
    pdf.set_xy(15, y_evt + 1)
    pdf.set_font("Helvetica", "B", 9)
    pdf.set_text_color(*C_CORE)
    pdf.cell(262, 8, "EVENT BUS: InMemory (Dev) / NATS JetStream (Prod)  -  Domain Events + Vernon Sync Triggers", align="C")

    # Arrows down to data
    pdf.draw_arrow(80, y_evt + 10, 80, y_evt + 17, C_CORE)
    pdf.draw_arrow(148, y_evt + 10, 148, y_evt + 17, C_CORE)
    pdf.draw_arrow(215, y_evt + 10, 215, y_evt + 17, C_CORE)

    # Data layer
    y_data = y_evt + 18
    pdf.draw_rounded_box(15, y_data, 80, 22, C_BG_GREEN, C_CORE)
    pdf.set_xy(15, y_data + 2)
    pdf.set_font("Helvetica", "B", 9)
    pdf.set_text_color(*C_CORE)
    pdf.cell(80, 5, "PostgreSQL", align="C")
    pdf.set_font("Helvetica", "", 7.5)
    pdf.set_text_color(*C_TEXT)
    pdf.set_xy(15, y_data + 8)
    pdf.cell(80, 4, "UUID v7 PK | JSONB (_rels/_data)", align="C")
    pdf.set_xy(15, y_data + 13)
    pdf.cell(80, 4, "Row-Level Tenant Isolation", align="C")

    pdf.draw_rounded_box(108, y_data, 80, 22, (255, 235, 238), C_RED)
    pdf.set_xy(108, y_data + 2)
    pdf.set_font("Helvetica", "B", 9)
    pdf.set_text_color(*C_RED)
    pdf.cell(80, 5, "Redis", align="C")
    pdf.set_font("Helvetica", "", 7.5)
    pdf.set_text_color(*C_TEXT)
    pdf.set_xy(108, y_data + 8)
    pdf.cell(80, 4, "Session Cache | Rate Limit", align="C")
    pdf.set_xy(108, y_data + 13)
    pdf.cell(80, 4, "Vernon Read-Cache Invalidation", align="C")

    pdf.draw_rounded_box(200, y_data, 77, 22, C_BG_LIGHT, C_INFRA)
    pdf.set_xy(200, y_data + 2)
    pdf.set_font("Helvetica", "B", 9)
    pdf.set_text_color(*C_INFRA)
    pdf.cell(77, 5, "Infrastructure", align="C")
    pdf.set_font("Helvetica", "", 7.5)
    pdf.set_text_color(*C_TEXT)
    pdf.set_xy(200, y_data + 8)
    pdf.cell(77, 4, "Uber FX (DI) | OpenTelemetry", align="C")
    pdf.set_xy(200, y_data + 13)
    pdf.cell(77, 4, "Docker | CI/CD Pipeline", align="C")

    # ═══════════════════════════════════════════════════════════════════════
    # PAGE 3: CLEAN ARCHITECTURE + CQRS FLOW
    # ═══════════════════════════════════════════════════════════════════════
    pdf.add_page()
    pdf.section_title("2. Clean Architecture + CQRS Flow")

    # Left side: Write path (Command)
    y0 = 30
    pdf.set_font("Helvetica", "B", 11)
    pdf.set_text_color(*C_RED)
    pdf.set_xy(15, y0)
    pdf.cell(130, 7, "COMMAND (Write Path)")

    pdf.set_font("Helvetica", "B", 11)
    pdf.set_text_color(*C_CORE)
    pdf.set_xy(155, y0)
    pdf.cell(130, 7, "QUERY (Read Path)")

    # Command flow boxes
    cmd_steps = [
        ("HTTP Request", "POST /api/v1/students", C_INFRA, C_BG_LIGHT),
        ("Handler (Layer 3)", "Validate request, extract JWT scope\nCall command handler", C_ACCENT, C_BG_ORANGE),
        ("Command Handler (Layer 2)", "Business logic, domain rules\nCreate entity, emit domain event", C_PRIMARY, C_BG_BOX),
        ("Repository (Layer 3)", "INSERT INTO students (...)\nPopulate _rels JSONB", C_CORE, C_BG_GREEN),
        ("Event Bus", "StudentCreated event published\nSyncEngine populates _data", (200, 80, 80), (255, 235, 238)),
    ]

    y = y0 + 10
    for i, (title, desc, border, bg) in enumerate(cmd_steps):
        pdf.draw_rounded_box(15, y, 130, 18, bg, border)
        pdf.set_xy(17, y + 1)
        pdf.set_font("Helvetica", "B", 8)
        pdf.set_text_color(*border)
        pdf.cell(126, 5, title)
        pdf.set_xy(17, y + 6)
        pdf.set_font("Helvetica", "", 7)
        pdf.set_text_color(*C_TEXT)
        pdf.multi_cell(126, 3.5, desc)
        if i < len(cmd_steps) - 1:
            pdf.draw_arrow(80, y + 18, 80, y + 23, border)
        y += 23

    # Query flow boxes
    qry_steps = [
        ("HTTP Request", "GET /api/v1/students?page=1", C_INFRA, C_BG_LIGHT),
        ("Handler (Layer 3)", "Validate query params, extract JWT scope\nCall query handler", C_ACCENT, C_BG_ORANGE),
        ("Query Handler (Layer 2)", "Build filter from scope\nReturn flat DTO (no domain logic)", C_PRIMARY, C_BG_BOX),
        ("Query Repository (Layer 3)", "SELECT id, name, _data FROM students\nZERO JOINs - read from _data JSONB", C_CORE, C_BG_GREEN),
        ("Response", "JSON with embedded _data\nclass_name, teacher_name already in row", (200, 80, 80), (255, 235, 238)),
    ]

    y = y0 + 10
    for i, (title, desc, border, bg) in enumerate(qry_steps):
        pdf.draw_rounded_box(155, y, 130, 18, bg, border)
        pdf.set_xy(157, y + 1)
        pdf.set_font("Helvetica", "B", 8)
        pdf.set_text_color(*border)
        pdf.cell(126, 5, title)
        pdf.set_xy(157, y + 6)
        pdf.set_font("Helvetica", "", 7)
        pdf.set_text_color(*C_TEXT)
        pdf.multi_cell(126, 3.5, desc)
        if i < len(qry_steps) - 1:
            pdf.draw_arrow(220, y + 18, 220, y + 23, border)
        y += 23

    # Key insight box
    y_note = y + 5
    pdf.draw_rounded_box(15, y_note, 270, 18, C_BG_ORANGE, C_ACCENT)
    pdf.set_xy(17, y_note + 2)
    pdf.set_font("Helvetica", "B", 9)
    pdf.set_text_color(*C_ACCENT)
    pdf.cell(266, 5, "KEY INSIGHT: Vernon Pattern eliminates JOINs on read path")
    pdf.set_xy(17, y_note + 7)
    pdf.set_font("Helvetica", "", 8)
    pdf.set_text_color(*C_TEXT)
    pdf.multi_cell(266, 4, "Write path: normal INSERT/UPDATE + populate _rels with FK IDs. SyncEngine (event-driven) populates _data with denormalized snapshots.\nRead path: SELECT from single table, _data JSONB already contains related entity names/codes. Result: <100ms for complex listing queries.")

    # ═══════════════════════════════════════════════════════════════════════
    # PAGE 4: MULTI-TENANT & AUTH FLOW
    # ═══════════════════════════════════════════════════════════════════════
    pdf.add_page()
    pdf.section_title("3. Multi-Tenant & Authentication Flow")

    pdf.sub_title("A. Four-Level Scope Hierarchy")
    y0 = 38

    levels = [
        ("TENANT", "Yayasan / Foundation (top-level SaaS customer)", C_PRIMARY, 15, 262),
        ("COMPANY", "Sekolah / Koperasi (business unit under tenant)", C_SECONDARY, 30, 232),
        ("BRANCH", "Cabang / Kampus (physical location)", C_ACCENT, 45, 202),
        ("WAREHOUSE", "Gudang / Unit Operasional (storage/ops)", C_KOPERASI, 60, 172),
    ]

    for label, desc, color, x_offset, width in levels:
        pdf.draw_rounded_box(x_offset, y0, width, 14, (*color, ), border_color=None)
        pdf.set_xy(x_offset, y0 + 1)
        pdf.set_font("Helvetica", "B", 10)
        pdf.set_text_color(*C_WHITE)
        pdf.cell(width, 6, label, align="C")
        pdf.set_xy(x_offset, y0 + 7)
        pdf.set_font("Helvetica", "", 7.5)
        pdf.cell(width, 5, desc, align="C")
        y0 += 17

    # Two-Phase JWT
    pdf.ln(5)
    y_jwt = 110
    pdf.sub_title("B. Two-Phase JWT Authentication")

    # Phase 1
    pdf.draw_rounded_box(15, y_jwt, 130, 30, C_BG_BOX, C_PRIMARY)
    pdf.set_xy(17, y_jwt + 1)
    pdf.set_font("Helvetica", "B", 9)
    pdf.set_text_color(*C_PRIMARY)
    pdf.cell(126, 5, "Phase 1: Identity Token")
    pdf.set_font("Helvetica", "", 7.5)
    pdf.set_text_color(*C_TEXT)
    phase1_text = "1. User login (email + password)\n2. Server returns Identity JWT\n3. Contains: user_id, email, tenant_list[]\n4. NO scope selected yet - user picks tenant"
    pdf.set_xy(17, y_jwt + 7)
    pdf.multi_cell(126, 4, phase1_text)

    pdf.draw_arrow(145, y_jwt + 15, 155, y_jwt + 15, C_PRIMARY)

    # Phase 2
    pdf.draw_rounded_box(157, y_jwt, 130, 30, C_BG_GREEN, C_CORE)
    pdf.set_xy(159, y_jwt + 1)
    pdf.set_font("Helvetica", "B", 9)
    pdf.set_text_color(*C_CORE)
    pdf.cell(126, 5, "Phase 2: Scoped Token")
    pdf.set_font("Helvetica", "", 7.5)
    pdf.set_text_color(*C_TEXT)
    phase2_text = "1. User selects tenant + company + branch\n2. Server returns Scoped JWT\n3. Contains: tenant_id, company_id, branch_id, roles[]\n4. ALL queries auto-filtered by scope"
    pdf.set_xy(159, y_jwt + 7)
    pdf.multi_cell(126, 4, phase2_text)

    # Dual mode
    y_dual = y_jwt + 38
    pdf.sub_title("C. Dual-Mode Institution Type")

    pdf.draw_rounded_box(15, y_dual + 8, 130, 25, C_BG_BOX, C_SEKOLAH)
    pdf.set_xy(17, y_dual + 9)
    pdf.set_font("Helvetica", "B", 9)
    pdf.set_text_color(*C_SEKOLAH)
    pdf.cell(126, 5, 'school_type: "general" | "islamic"')
    pdf.set_font("Helvetica", "", 7.5)
    pdf.set_text_color(*C_TEXT)
    pdf.set_xy(17, y_dual + 15)
    pdf.multi_cell(126, 4, "General: SD/SMP/SMA/SMK standard\nIslamic: Pondok Pesantren (Santri, Halaqah, Tahfiz)")

    pdf.draw_rounded_box(157, y_dual + 8, 130, 25, C_BG_PURPLE, C_KOPERASI)
    pdf.set_xy(159, y_dual + 9)
    pdf.set_font("Helvetica", "B", 9)
    pdf.set_text_color(*C_KOPERASI)
    pdf.cell(126, 5, 'coop_type: "general" | "islamic"')
    pdf.set_font("Helvetica", "", 7.5)
    pdf.set_text_color(*C_TEXT)
    pdf.set_xy(159, y_dual + 15)
    pdf.multi_cell(126, 4, "General: Koperasi konvensional (bunga)\nIslamic: BMT (bagi hasil, akad syariah, zakat)")

    # Policy note
    y_pol = y_dual + 38
    pdf.draw_rounded_box(15, y_pol, 272, 14, C_BG_ORANGE, C_ACCENT)
    pdf.set_xy(17, y_pol + 2)
    pdf.set_font("Helvetica", "B", 8)
    pdf.set_text_color(*C_ACCENT)
    pdf.cell(268, 4, "POLICY: Immutable After Activation")
    pdf.set_xy(17, y_pol + 6)
    pdf.set_font("Helvetica", "", 7.5)
    pdf.set_text_color(*C_TEXT)
    pdf.cell(268, 5, "school_type & coop_type are set ONCE during onboarding and CANNOT be changed. Both fields are independent (4 possible combinations).")

    # ═══════════════════════════════════════════════════════════════════════
    # PAGE 5: VERNON PATTERN DETAIL
    # ═══════════════════════════════════════════════════════════════════════
    pdf.add_page()
    pdf.section_title("4. Vernon Denormalized Read-Cache Pattern")

    y0 = 30
    pdf.sub_title("How _rels and _data work")

    # Table illustration
    pdf.draw_rounded_box(15, y0 + 8, 270, 55, C_BG_LIGHT, C_INFRA)
    pdf.set_xy(17, y0 + 10)
    pdf.set_font("Courier", "B", 8)
    pdf.set_text_color(*C_TEXT)

    table_text = """TABLE: students
+------------+----------+----------+---------+---------------------------+----------------------------------------+
| id (UUID7) | name     | class_id | ...     | _rels (JSONB)             | _data (JSONB)                          |
+------------+----------+----------+---------+---------------------------+----------------------------------------+
| abc-123    | Ahmad    | cls-01   | ...     | {"class_id":"cls-01",     | {"class":{"name":"7A","level":"VII"},  |
|            |          |          |         |  "teacher_id":"tch-05",   |  "teacher":{"name":"Pak Budi"},        |
|            |          |          |         |  "academic_year":"ay-02"} |  "academic_year":{"name":"2025/2026"}} |
+------------+----------+----------+---------+---------------------------+----------------------------------------+"""

    for line in table_text.strip().split('\n'):
        pdf.set_xy(17, pdf.get_y() + 3.5)
        pdf.cell(266, 3, line)

    # Sync flow
    y_sync = y0 + 68
    pdf.sub_title("Sync Engine Flow")

    sync_steps = [
        ("1. Write", "Class '7A' name changed\nto '7B'", C_PRIMARY),
        ("2. Event", "ClassUpdated event\npublished to bus", C_ACCENT),
        ("3. SyncEngine", "Finds all rows where\n_rels.class_id = cls-01", C_CORE),
        ("4. Update _data", "Updates _data.class.name\nfrom '7A' to '7B'", C_KOPERASI),
    ]

    x = 15
    for i, (title, desc, color) in enumerate(sync_steps):
        pdf.draw_rounded_box(x, y_sync + 5, 62, 22, C_WHITE, color)
        pdf.set_xy(x + 2, y_sync + 6)
        pdf.set_font("Helvetica", "B", 8)
        pdf.set_text_color(*color)
        pdf.cell(58, 5, title)
        pdf.set_xy(x + 2, y_sync + 12)
        pdf.set_font("Helvetica", "", 7)
        pdf.set_text_color(*C_TEXT)
        pdf.multi_cell(58, 3.5, desc)
        if i < len(sync_steps) - 1:
            pdf.draw_arrow(x + 62, y_sync + 16, x + 70, y_sync + 16, color)
        x += 69

    # Benefits
    y_ben = y_sync + 35
    pdf.sub_title("Benefits & Trade-offs")

    pdf.draw_rounded_box(15, y_ben + 5, 130, 35, C_BG_GREEN, C_CORE)
    pdf.set_xy(17, y_ben + 7)
    pdf.set_font("Helvetica", "B", 9)
    pdf.set_text_color(*C_CORE)
    pdf.cell(126, 5, "BENEFITS")
    pdf.set_font("Helvetica", "", 7.5)
    pdf.set_text_color(*C_TEXT)
    benefits = "+ Zero JOINs on read path (<100ms)\n+ Schema-flexible (new fields = JSONB update, no ALTER)\n+ Read:Write asymmetry optimized (10:1 to 100:1)\n+ Single-table queries = simple pagination & filtering\n+ Consistent denormalization via SyncEngine"
    pdf.set_xy(17, y_ben + 13)
    pdf.multi_cell(126, 4, benefits)

    pdf.draw_rounded_box(155, y_ben + 5, 130, 35, (255, 235, 238), C_RED)
    pdf.set_xy(157, y_ben + 7)
    pdf.set_font("Helvetica", "B", 9)
    pdf.set_text_color(*C_RED)
    pdf.cell(126, 5, "TRADE-OFFS")
    pdf.set_font("Helvetica", "", 7.5)
    pdf.set_text_color(*C_TEXT)
    tradeoffs = "- Write path slightly slower (populate _rels)\n- Storage overhead (JSONB per row)\n- Eventual consistency (_data lags by event latency)\n- SyncEngine must handle cascading updates\n- Need GIN indexes on _rels and _data"
    pdf.set_xy(157, y_ben + 13)
    pdf.multi_cell(126, 4, tradeoffs)

    # ═══════════════════════════════════════════════════════════════════════
    # PAGE 6: POLICIES & RULES
    # ═══════════════════════════════════════════════════════════════════════
    pdf.add_page()
    pdf.section_title("5. System Policies & Business Rules")

    policies = [
        ("Multi-Tenant Data Isolation", C_PRIMARY, C_BG_BOX, [
            "Every table has tenant_id - enforced at middleware level",
            "Row-Level Security: queries auto-filtered by JWT scope (tenant/company/branch)",
            "Cross-tenant access only for superadmin with explicit scope override",
            "Operational tables have full 4-level hierarchy (tenant > company > branch > warehouse)",
        ]),
        ("RBAC & Authorization", C_ACCENT, C_BG_ORANGE, [
            "Role-based access control with hierarchical permissions",
            "Roles: SuperAdmin, Admin Yayasan, Kepala Sekolah, Guru, TU, Bendahara, Orang Tua, Siswa",
            "Koperasi roles: Manager, Teller, Nasabah, Auditor",
            "Permission inheritance: Yayasan roles can access all schools under their tenant",
        ]),
        ("Financial Transaction Policies (Koperasi)", C_KOPERASI, C_BG_PURPLE, [
            "Double-entry accounting (every transaction = debit + credit journal)",
            "Transaction immutability - no UPDATE/DELETE, only reversal transactions",
            "Teller session: cash must balance at session close (open -> transact -> close -> reconcile)",
            "Multi-level approval for loans (credit scoring + committee approval)",
            "NPL classification: Lancar > DPK > Kurang Lancar > Diragukan > Macet",
            "Regulatory compliance: OJK/LKM reporting, UU PDP (data privacy), tax integration",
        ]),
        ("Academic Policies (Sekolah)", C_SEKOLAH, C_BG_BOX, [
            "Academic Year as universal time scope - all records tied to tahun_ajaran_id",
            "Student Class Placement: one active placement per student per semester",
            "Attendance: daily + per-subject tracking with status (Hadir/Sakit/Izin/Alpha)",
            "Rapor generation: orchestrates grades + attendance + extracurricular + discipline",
            "Dapodik integration: sync student/teacher data to national education database",
        ]),
        ("Dual-Mode Business Rules", C_CORE, C_BG_GREEN, [
            "Conventional: interest-based (bunga), standard koperasi terms",
            "Islamic (BMT): profit-sharing (bagi hasil/nisbah), syariah contracts (akad)",
            "BMT-specific: Zakat, Infaq, Ta'zir (social fund penalty), Rahn (collateral)",
            "Label mapping: terminology auto-switches based on institution type config",
            "Type is IMMUTABLE after tenant activation - prevents data inconsistency",
        ]),
        ("Event-Driven Architecture", C_INFRA, C_BG_LIGHT, [
            "Domain events published via Event Bus (InMemory for dev, NATS JetStream for prod)",
            "Vernon SyncEngine listens for entity changes and updates _data JSONB",
            "At-least-once delivery guarantee with idempotent handlers",
            "Dead-letter queue for failed event processing with retry + alerting",
        ]),
    ]

    y = 30
    for title, color, bg, items in policies:
        h = 8 + len(items) * 5
        if y + h > 190:
            pdf.add_page()
            pdf.section_title("5. System Policies & Business Rules (cont.)")
            y = 30
        pdf.draw_rounded_box(15, y, 270, h, bg, color)
        pdf.set_xy(17, y + 1)
        pdf.set_font("Helvetica", "B", 9)
        pdf.set_text_color(*color)
        pdf.cell(266, 5, title)
        iy = y + 7
        for item in items:
            pdf.set_xy(20, iy)
            pdf.set_font("Helvetica", "", 7.5)
            pdf.set_text_color(*C_TEXT)
            pdf.cell(3, 4, "-")
            pdf.cell(260, 4, "  " + item)
            iy += 5
        y += h + 4

    # ═══════════════════════════════════════════════════════════════════════
    # PAGE 7: DOMAIN MAP
    # ═══════════════════════════════════════════════════════════════════════
    pdf.add_page()
    pdf.section_title("6. Domain Map - All Modules")

    # Sekolah domains
    y0 = 30
    pdf.draw_rounded_box(15, y0, 270, 8, C_SEKOLAH)
    pdf.set_xy(15, y0 + 1)
    pdf.set_font("Helvetica", "B", 10)
    pdf.set_text_color(*C_WHITE)
    pdf.cell(270, 6, "MANAGEMENT SEKOLAH - 58 ADR across 13 domain groups", align="C")

    sekolah_domains = [
        ("Siswa", "S001-S018", "Data siswa, wali, akademik, kesehatan, keuangan, disiplin, ekskul, PPDB, BK, rapor"),
        ("Kurikulum", "S019-S025", "Kurikulum, mata pelajaran, jadwal, ujian, kalender, RPP, jurnal mengajar"),
        ("Guru & HR", "S026-S032", "Absensi guru, beban mengajar, PKG, PKB, cuti, payroll, piket"),
        ("Asrama", "S033-S035", "Manajemen kamar, aktivitas & absensi asrama, disiplin & kesehatan"),
        ("Kantin", "S036-S037", "Menu & vendor management, transaksi & billing kantin"),
        ("Fasilitas", "S038-S041", "Perpustakaan, laboratorium, aset & inventaris, booking ruangan"),
        ("Komunikasi", "S042-S045", "Portal orang tua, messaging, notifikasi multi-channel, pengumuman"),
        ("Administrasi", "S046-S049", "Surat menyurat, approval workflow, profil sekolah, komite"),
        ("Keuangan", "S050-S051", "RKAS (anggaran), payment gateway"),
        ("Transport", "S052", "Manajemen antar-jemput siswa"),
        ("Integrasi", "S053-S055", "E-learning, reporting & analytics, Dapodik sync"),
        ("Alumni", "S056-S058", "Alumni, beasiswa, kegiatan sekolah"),
    ]

    y = y0 + 10
    for name, adr_range, desc in sekolah_domains:
        pdf.set_xy(17, y)
        pdf.set_font("Helvetica", "B", 7.5)
        pdf.set_text_color(*C_SEKOLAH)
        pdf.cell(25, 4, name)
        pdf.set_font("Courier", "", 6.5)
        pdf.set_text_color(*C_TEXT_LIGHT)
        pdf.cell(22, 4, adr_range)
        pdf.set_font("Helvetica", "", 7)
        pdf.set_text_color(*C_TEXT)
        pdf.cell(220, 4, desc)
        y += 5.5

    # Koperasi domains
    y += 5
    pdf.draw_rounded_box(15, y, 270, 8, C_KOPERASI)
    pdf.set_xy(15, y + 1)
    pdf.set_font("Helvetica", "B", 10)
    pdf.set_text_color(*C_WHITE)
    pdf.cell(270, 6, "MANAGEMENT KOPERASI SEKOLAH - 24 ADR across 7 domain groups", align="C")

    koperasi_domains = [
        ("Core", "K001-K003", "Nasabah/anggota, rekening, produk & akad (konvensional + syariah)"),
        ("Simpanan", "K004-K006", "Simpanan pokok/wajib, tabungan multi-produk, deposito/berjangka"),
        ("Pembiayaan", "K007-K010", "Pinjaman/pembiayaan, angsuran, denda/penalti, jaminan/agunan"),
        ("Transaksi", "K011-K014", "Transaction engine, teller session, denominasi uang, kas & cash flow"),
        ("Akuntansi", "K015-K018", "Jurnal & COA, SHU, laporan regulasi (OJK/Dinas), zakat & infaq (BMT)"),
        ("Extensions", "K019-K021", "Toko & kantin POS, payroll auto-deduction, e-wallet/uang saku digital"),
        ("Platform", "K022-K024", "Notifikasi, dashboard & portal, integrasi & API"),
    ]

    y += 10
    for name, adr_range, desc in koperasi_domains:
        pdf.set_xy(17, y)
        pdf.set_font("Helvetica", "B", 7.5)
        pdf.set_text_color(*C_KOPERASI)
        pdf.cell(25, 4, name)
        pdf.set_font("Courier", "", 6.5)
        pdf.set_text_color(*C_TEXT_LIGHT)
        pdf.cell(22, 4, adr_range)
        pdf.set_font("Helvetica", "", 7)
        pdf.set_text_color(*C_TEXT)
        pdf.cell(220, 4, desc)
        y += 5.5

    # Core domains
    y += 5
    pdf.draw_rounded_box(15, y, 270, 8, C_CORE)
    pdf.set_xy(15, y + 1)
    pdf.set_font("Helvetica", "B", 10)
    pdf.set_text_color(*C_WHITE)
    pdf.cell(270, 6, "CORE (Shared Foundation) - 18 ADR", align="C")

    core_domains = [
        ("Architecture", "ADR-001..009", "Clean Arch + CQRS, Vernon Pattern, Event Bus, DI, UUID v7, Multi-Tenant, Dual-Mode"),
        ("Master Data", "ADR-010..013", "Tahun Ajaran, Kelas & Rombel, Guru & Staff, Users & Roles (RBAC)"),
        ("Infrastructure", "ADR-014..018", "Sync Engine, Frontend Multi-App, CI/CD, Data Migration, Regulatory Compliance"),
        ("Business", "BIZ-001", "Business Model, Pricing Strategy, Unit Economics"),
    ]

    y += 10
    for name, adr_range, desc in core_domains:
        pdf.set_xy(17, y)
        pdf.set_font("Helvetica", "B", 7.5)
        pdf.set_text_color(*C_CORE)
        pdf.cell(25, 4, name)
        pdf.set_font("Courier", "", 6.5)
        pdf.set_text_color(*C_TEXT_LIGHT)
        pdf.cell(25, 4, adr_range)
        pdf.set_font("Helvetica", "", 7)
        pdf.set_text_color(*C_TEXT)
        pdf.cell(217, 4, desc)
        y += 5.5

    # ═══════════════════════════════════════════════════════════════════════
    # PAGE 8: SPRINT ROADMAP
    # ═══════════════════════════════════════════════════════════════════════
    pdf.add_page()
    pdf.section_title("7. Sprint Roadmap - 15 Phases")

    sprints = [
        ("Sprint 0", "Fase 0", "Platform Foundation", "8 ADR", "Clean Arch, CQRS, Vernon, Event Bus, DI, UUID v7, React, Dual-Mode", C_CORE),
        ("Sprint 1", "Fase 1", "Auth & Multi-Tenant", "2 ADR", "4-Level Hierarchy, Two-Phase JWT, RBAC, User CRUD", C_PRIMARY),
        ("Sprint 2", "Fase 2+3", "Master Data + Core Koperasi", "6 ADR", "Tahun Ajaran, Guru, Kelas + Produk, Nasabah, Rekening", C_SECONDARY),
        ("Sprint 3", "Fase 4A+4B", "Student Core + Simpanan", "10 ADR", "Data Siswa, Wali, PPDB + Simpanan Pokok/Wajib, Tabungan, Deposito", C_SEKOLAH),
        ("Sprint 4", "Fase 5A+5B", "Kurikulum + Pembiayaan", "10 ADR", "Kurikulum, Jadwal, RPP + Pinjaman, Angsuran, Denda, Jaminan", C_ACCENT),
        ("Sprint 5", "Fase 6A+6B", "Kehadiran & Nilai + Transaksi", "12 ADR", "Absensi, Nilai, Ujian, Ekskul + Transaction Engine, Teller, Kas", C_RED),
        ("Sprint 6", "Fase 7A+7B", "Rapor & HR + Akuntansi", "11 ADR", "Rapor, SPP, HR Guru + Jurnal COA, SHU, Laporan OJK, Zakat", C_KOPERASI),
        ("Sprint 7", "Fase 8", "Fasilitas & Sarana", "7 ADR", "Perpustakaan, Lab, Aset, Booking, PKG, PKB, Payroll", C_INFRA),
        ("Sprint 8", "Fase 9+10", "Asrama/Kantin + Ext.", "8 ADR", "Dormitory, Canteen + Toko POS, E-Wallet, Auto-Deduction", C_ACCENT),
        ("Sprint 9", "Fase 11", "Komunikasi & Portal", "7 ADR", "Notifikasi, Messaging, Portal Orang Tua, Dashboard", C_SEKOLAH),
        ("Sprint 10", "Fase 12", "Administrasi & Keuangan", "6 ADR", "Surat TU, Approval, Profil Sekolah, RKAS, Payment Gateway", C_PRIMARY),
        ("Sprint 11", "Fase 13", "Integrasi & Pelaporan", "5 ADR", "Analytics, Dapodik, E-Learning, Transport, API Koperasi", C_CORE),
        ("Sprint 12+", "Fase 14", "Nice-to-Have", "3 ADR", "Alumni, Beasiswa, Kegiatan Sekolah", C_TEXT_LIGHT),
        ("Ongoing", "Fase INF", "Infrastructure", "6 ADR", "Sync Engine, Frontend Arch, CI/CD, Migration, Compliance, Pricing", C_INFRA),
    ]

    y = 30
    # Header row
    pdf.set_font("Helvetica", "B", 7.5)
    pdf.set_text_color(*C_WHITE)
    pdf.draw_rounded_box(15, y, 270, 7, C_TEXT)
    pdf.set_xy(16, y + 1)
    pdf.cell(28, 5, "Sprint")
    pdf.cell(22, 5, "Phase")
    pdf.cell(52, 5, "Focus")
    pdf.cell(18, 5, "ADRs")
    pdf.cell(150, 5, "Key Deliverables")
    y += 9

    for sprint, phase, focus, adr_count, deliverables, color in sprints:
        bg = (*color[0:3],) if color != C_TEXT_LIGHT else C_BG_LIGHT
        # Lighter bg
        light_bg = tuple(min(255, int(c + (255 - c) * 0.85)) for c in bg)
        pdf.draw_rounded_box(15, y, 270, 9, light_bg, color)
        pdf.set_xy(16, y + 1.5)
        pdf.set_font("Helvetica", "B", 7)
        pdf.set_text_color(*color)
        pdf.cell(28, 5, sprint)
        pdf.set_font("Helvetica", "", 7)
        pdf.set_text_color(*C_TEXT)
        pdf.cell(22, 5, phase)
        pdf.set_font("Helvetica", "B", 7)
        pdf.cell(52, 5, focus)
        pdf.set_font("Courier", "", 6.5)
        pdf.set_text_color(*C_TEXT_LIGHT)
        pdf.cell(18, 5, adr_count)
        pdf.set_font("Helvetica", "", 6.5)
        pdf.set_text_color(*C_TEXT)
        pdf.cell(150, 5, deliverables)
        y += 11

    # Total
    y += 3
    pdf.draw_rounded_box(15, y, 270, 10, C_BG_ORANGE, C_ACCENT)
    pdf.set_xy(17, y + 2)
    pdf.set_font("Helvetica", "B", 9)
    pdf.set_text_color(*C_ACCENT)
    pdf.cell(266, 6, "TOTAL: 100 ADRs  |  Core: 18  +  Sekolah: 58  +  Koperasi: 24  |  ~100+ Database Tables  |  16 Stakeholder Roles", align="C")

    # ═══════════════════════════════════════════════════════════════════════
    # PAGE 9: STAKEHOLDER MAP
    # ═══════════════════════════════════════════════════════════════════════
    pdf.add_page()
    pdf.section_title("8. Stakeholder Coverage")

    stakeholders = [
        ("Kepala Sekolah", "Strategic oversight, PKG approval, analytics dashboard, accreditation", C_PRIMARY),
        ("Wakasek Kurikulum", "Curriculum, subject management, schedule approval, RPP review", C_SEKOLAH),
        ("Wakasek Kesiswaan", "Student discipline, extracurricular, counseling, events", C_SEKOLAH),
        ("Wakasek Sarana", "Library, laboratory, asset inventory, facility booking", C_SEKOLAH),
        ("Guru / Teacher", "Attendance input, grade input, lesson plan, teaching journal, workload", C_SECONDARY),
        ("Tata Usaha (TU)", "Correspondence, student documents, school profile", C_INFRA),
        ("Bendahara", "SPP collection, RKAS budget, payment gateway, payroll", C_ACCENT),
        ("Orang Tua / Wali", "Parent portal, notifications, messaging, student progress", C_CORE),
        ("Siswa / Student", "Dashboard, academic records, e-wallet, attendance view", C_SEKOLAH),
        ("Manager Koperasi", "All koperasi operations, loan approval, SHU, reporting", C_KOPERASI),
        ("Teller Koperasi", "Transaction processing, cash management, daily reconciliation", C_KOPERASI),
        ("Nasabah Koperasi", "Self-service portal, savings, loan status, statements", C_KOPERASI),
        ("Dinas Pendidikan", "Dapodik integration, school reporting, regulatory compliance", C_PRIMARY),
        ("OJK / Regulator", "Financial reports, NPL ratios, CAR, compliance reports", C_RED),
        ("Pustakawan/Laboran", "Library/lab management, booking, inventory", C_INFRA),
        ("Komite Sekolah", "Budget approval, school oversight, committee management", C_ACCENT),
    ]

    y = 30
    col1_w = 135
    col2_x = 152

    for i, (role, scope, color) in enumerate(stakeholders):
        x = 15 if i % 2 == 0 else col2_x
        if i % 2 == 0 and i > 0:
            y += 12
        if y > 185:
            pdf.add_page()
            y = 20
        pdf.draw_rounded_box(x, y, col1_w, 10, C_WHITE, color)
        pdf.set_xy(x + 2, y + 0.5)
        pdf.set_font("Helvetica", "B", 7.5)
        pdf.set_text_color(*color)
        pdf.cell(30, 4, role)
        pdf.set_xy(x + 2, y + 4.5)
        pdf.set_font("Helvetica", "", 6.5)
        pdf.set_text_color(*C_TEXT)
        pdf.cell(col1_w - 4, 4, scope)

    # ═══════════════════════════════════════════════════════════════════════
    # SAVE
    # ═══════════════════════════════════════════════════════════════════════
    pdf.output(OUTPUT_PATH)
    print(f"PDF saved to: {OUTPUT_PATH}")
    print(f"Pages: {pdf.page_no()}")


if __name__ == "__main__":
    build_pdf()
