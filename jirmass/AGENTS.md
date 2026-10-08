# Agent Guidelines for Jirmass

Welcome to the **Jirmass** repository. This document defines the architecture, conventions, workflows, and operational guardrails for AI coding agents and contributors working on this codebase.

---

## 1. Project Overview & Architecture

Jirmass is a focused web application for a manufacturing company specializing exclusively in **Water Packing Bags** and **Water Packing Rolls** (sachet water packaging film). It consists of two primary components:

1. **Go Web Application (Frontend / Main Server)**
   - Location: Project root (`./`)
   - Technology: Go 1.22+ (`net/http`, `html/template`)
   - Purpose: Renders customer-facing web pages (`/`, `/about`, `/contact`), provides direct order placement (`/api/order`), automated customer support bot (`/api/bot/chat`), and serves static assets.
   - Default Port: `:8081` (`http://localhost:8081`)

2. **Django REST API (Backend)**
   - Location: `backend/`
   - Technology: Python 3.12+ (Django & Django REST Framework)
   - Purpose: Manages water packaging orders, inquiries, notifications, and headless API endpoints alongside Django Admin (`/admin/`).
   - Database: SQLite (`backend/db.sqlite3`) for development.

---

## 2. Directory Layout

```text
jirmass/
├── go.mod                      # Go module definition (module jirmass, Go 1.22+)
├── main.go                     # Go server entrypoint, routes, direct order & page handlers
├── bot.go                      # Automated water packaging support bot & inquiry logger
├── main_test.go                # Unit and HTTP integration tests for Go server, bot & direct orders
├── static/                     # Static assets served by Go under /static/
│   ├── css/
│   │   └── style.css           # Global stylesheet with responsive layout & direct order form
│   ├── js/
│   │   ├── bot.js              # Client-side automated chatbot widget controller
│   │   └── order.js            # Client-side direct order placement submission controller
│   └── images/                 # Brand logos and industrial background graphics
│       ├── README.md           # Instructions for inserting custom images
│       ├── logo.svg            # Primary brand vector logo placeholder
│       ├── logo-icon.svg       # Standalone logo mark / favicon placeholder
│       ├── hero-bg.svg         # Hero section industrial mesh vector backdrop placeholder
│       └── pattern-bg.svg      # Page background hexagonal polymer lattice pattern
├── templates/                  # Go HTML templates
│   ├── home.html               # Homepage with water packaging showcase & Direct Order section
│   ├── about.html              # About manufacturing standards template
│   └── contact.html            # Contact & sales inquiry submission template
├── backend/                    # Django application root
│   ├── manage.py               # Django management script
│   ├── requirements.txt        # Backend Python dependencies
│   ├── config/                 # Django project settings & root routing
│   │   ├── asgi.py             # ASGI config
│   │   ├── settings.py         # Django settings (apps, DB, middleware)
│   │   ├── urls.py             # Root URL patterns (/admin/, /api/products/)
│   │   └── wsgi.py             # WSGI config
│   └── products/               # Orders, Inquiries, and Brand Customization Django app
│       ├── admin.py            # Model admin registration (Orders, Inquiries, Brand Assets)
│       ├── apps.py             # AppConfig
│       ├── migrations/         # Django schema migrations
│       ├── models.py           # Order, CustomerInquiry, BrandAsset, and AdminAvailability models
│       ├── serializers.py      # Django REST Framework serializers
│       ├── tests.py            # Unit and API integration tests
│       ├── urls.py             # App URL routes (/api/products/orders/, /inquiries/, etc.)
│       └── views.py            # DRF API views
└── AGENTS.md                   # This instruction file
```

---

## 3. Development Setup & Key Commands

### Go Web Server (Root Directory)

Always execute Go commands from the project root:

- **Run Server:**
  ```bash
  go run .
  ```
- **Build / Compile Verification:**
  ```bash
  go build .
  ```
  *(Important: If testing compilation produces a binary, delete the binary before finishing to avoid committing build artifacts).*
- **Run Tests:**
  ```bash
  go test -v ./...
  ```
- **Format Code:**
  ```bash
  go fmt ./...
  ```
- **Static Analysis / Vet:**
  ```bash
  go vet ./...
  ```

### Django Backend (`backend/` Directory)

Run Django commands either from within `backend/` or referencing `backend/manage.py`:

- **Environment Setup:**
  ```bash
  python3 -m venv .venv
  source .venv/bin/activate
  pip install django djangorestframework
  ```
- **Run Development Server:**
  ```bash
  cd backend && python manage.py runserver 8000
  ```
- **Apply Database Migrations:**
  ```bash
  cd backend && python manage.py migrate
  ```
- **Create Schema Migrations:**
  ```bash
  cd backend && python manage.py makemigrations
  ```
- **Run Django Tests:**
  ```bash
  cd backend && python manage.py test
  ```
- **System Check / Lint:**
  ```bash
  cd backend && python manage.py check
  ```

---

## 4. Coding Standards & Conventions

### Go Standards
- **Idiomatic Go:** Follow standard Go style guidelines and use `go fmt` / `go vet`.
- **Error Handling:** Check errors explicitly (`if err != nil`). Never discard errors silently. When returning HTTP error responses, use appropriate HTTP status codes (e.g., `http.StatusInternalServerError`, `http.StatusNotFound`).
- **Template Rendering:** Parse and render templates with care; handle missing template files gracefully.
- **Port Management:** The Go server defaults to `:8081`. Do not change default ports unless explicitly instructed.

### Python / Django Standards
- **PEP 8:** Follow standard Python formatting and naming conventions (snake_case for variables/functions, PascalCase for classes).
- **Django REST Framework:**
  - Ensure `rest_framework` is properly configured in `INSTALLED_APPS` in `backend/config/settings.py` when using DRF components.
  - Use generic API views (`generics.ListCreateAPIView`, `generics.RetrieveUpdateDestroyAPIView`) or ViewSets for API consistency.
  - Keep serializers in `serializers.py` and views in `views.py`.
- **Database & Migrations:**
  - Whenever modifying models in `backend/products/models.py`, generate new migrations using `makemigrations`.
  - Never edit existing applied migration files manually; generate sequential migrations.
- **Environment & Secrets:**
  - Do not hardcode secret keys, passwords, or production credentials in `backend/config/settings.py`. Use environment variables where appropriate.

### Frontend & Templates
- Keep HTML semantic and accessible.
- Maintain CSS styling within `static/css/style.css` or modular CSS files in `static/css/`.
- Ensure paths to static assets in templates use the `/static/` URL prefix.

---

## 5. Agent Workflow & Safety Rules

1. **Context Awareness:**
   - Determine whether a task pertains to the **Go frontend** (`./`) or the **Django backend** (`backend/`). Use the appropriate working directory for each tool.
2. **Pre-Completion Verification:**
   - For Go changes, run `go build .` and `go test ./...` to verify syntax and build integrity. Clean up any compiled binaries immediately.
   - For Django changes, run `python backend/manage.py check` to ensure syntax, dependencies, and settings are valid.
3. **Clean Git Tree:**
   - Never commit build binaries (`jirmass`, `*.exe`), database files (`*.sqlite3`), virtual environment folders (`.venv/`, `env/`), or cache folders (`__pycache__/`).
   - Respect `.gitignore` rules at all times.
4. **Preserve Documentation:**
   - Preserve existing docstrings, comments, and structure unless directly instructed to modify them.

---

## 6. Feature Configuration & Integration Reference

### Brand Assets & Background Graphics Customization
All visual brand assets are modular and served from `static/images/`:
- **Company Logo**: Replace [`static/images/logo.svg`](file:///home/l2e_abdyhaya/dev-1o1/jirmass/static/images/logo.svg) with your SVG or PNG vector logo. Used across the navigation header.
- **Logo Icon / Favicon**: Replace [`static/images/logo-icon.svg`](file:///home/l2e_abdyhaya/dev-1o1/jirmass/static/images/logo-icon.svg).
- **Hero Background Mesh**: Replace [`static/images/hero-bg.svg`](file:///home/l2e_abdyhaya/dev-1o1/jirmass/static/images/hero-bg.svg) for the main homepage banner.
- **Hexagonal Lattice Pattern**: Replace [`static/images/pattern-bg.svg`](file:///home/l2e_abdyhaya/dev-1o1/jirmass/static/images/pattern-bg.svg) for page backdrops.

### Live Admin Notifications (Email & Webhook Alerts)
When the customer bot or contact form registers an inquiry while the admin is away, or when a wholesale order is submitted:
- **Email Alerts**: Dispatched via Django's `send_mail` to the configured `admin_email` (customizable in Django Admin under **Admin Availability & Alerts Settings**).
- **Webhook Alerts**: Dispatched as an asynchronous HTTP POST with JSON event payload to the configured `webhook_url` (compatible with Slack, Discord, Zapier, WhatsApp webhook).
- **Environment Variable (Go)**: Set `ADMIN_WEBHOOK_URL="https://your-webhook-endpoint"` to enable instant Go-level webhook dispatch.

### LLM / AI Bot Integration
- **Google Gemini API**: Set `GEMINI_API_KEY="your-api-key"` (or `LLM_API_KEY`) and optionally `LLM_MODEL="gemini-1.5-flash"` in your environment.
- **Autonomous Fallback**: If no API key is provided or if network latency occurs, the bot seamlessly falls back to the deterministic local knowledge base.

### Direct Order Placement Flow
- **Direct Order Form**: Customer selects water packaging product type (Rolls, Bags, or Both), print customization (Plain or Custom Branded), quantities, company/factory name, contact details, and factory delivery address.
- **Order Endpoint**: `POST /api/order` validates submission, creates unique order reference `WTR-YYYYMMDD-XXXX`, dispatches live admin email and webhook alerts, and forwards order details to the Django backend.

