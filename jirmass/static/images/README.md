# Static Images & Brand Assets Directory

This folder is the designated storage location for all frontend brand assets and graphics.

---

## 1. Files To Insert

Replace the placeholder files in this folder with your actual images:

| Filename | Purpose | Recommended Format & Dimensions |
| :--- | :--- | :--- |
| `logo.svg` *(or `logo.png`)* | Navigation header brand logo | Vector SVG or transparent PNG (approx. 320×64 px) |
| `logo-icon.svg` *(or `favicon.ico`)* | Browser tab favicon and mobile mark | Vector SVG or 32×32 / 64×64 PNG/ICO |
| `hero-bg.svg` *(or `hero-bg.jpg`)* | Homepage banner background graphic | High-res SVG, JPG, or PNG (approx. 1440×600 px) |
| `pattern-bg.svg` *(or `pattern-bg.png`)* | Subtle page background pattern | Repeating tile pattern (optional) |

---

## 2. How To Insert Your Images

### Option A: Direct File Placement (Fastest)
Simply copy or drag-and-drop your image files into this directory (`static/images/`):
```bash
cp /path/to/your/actual_logo.svg static/images/logo.svg
cp /path/to/your/actual_hero.jpg static/images/hero-bg.jpg
```

### Option B: Using PNG or JPG Format
If your logo is a PNG file (e.g. `logo.png`) instead of SVG:
1. Place `logo.png` in this directory (`static/images/logo.png`).
2. The templates are equipped with automatic fallback detection, or you can update the image tag in `templates/*.html`:
   ```html
   <img src="/static/images/logo.png" alt="Company Logo" class="brand-logo-img">
   ```

### Option C: Upload Via Django Admin
Log in to Django Admin at `http://localhost:8000/admin/`, navigate to **Brand Assets & Graphics**, and upload your logo and background graphics directly through your browser.

---

## 3. Automatic Graceful Fallback
If an image file is absent or not yet inserted, the website automatically falls back to clean, responsive text badges and subtle dark slate gradients without displaying broken image icons.
