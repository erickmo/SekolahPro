# File Upload Components

Dokumentasi untuk komponen file upload di FormPageTemplate widgets.

## Komponen Tersedia

### 1. FileUploadField

Komponen input file dengan button click untuk memilih file.

**Props:**
```typescript
interface FileUploadFieldProps {
  value: string                    // URL file yang sudah diupload
  onChange: (url: string) => void  // Callback saat file berhasil diupload
  label: string                    // Label field
  hint?: string                    // Teks bantuan (opsional)
  required?: boolean               // Tandai field sebagai required
  accept?: string                  // MIME types yang diterima (default: 'image/*')
  maxSizeMB?: number              // Ukuran file max dalam MB (default: 10)
  previewImageUrl?: string        // URL preview image
}
```

**Contoh Penggunaan:**
```tsx
import { FileUploadField } from '@/widgets/FormPageTemplate'

export function MyForm() {
  const [logoUrl, setLogoUrl] = useState('')

  return (
    <FileUploadField
      label="Logo Perusahaan"
      hint="Format PNG, JPG, atau WebP"
      value={logoUrl}
      onChange={setLogoUrl}
      accept="image/png,image/jpeg,image/webp"
      maxSizeMB={5}
      previewImageUrl={logoUrl}
    />
  )
}
```

### 2. FileDropZone

Komponen drag & drop untuk upload file dengan visual feedback yang lebih interaktif.

**Props:**
```typescript
interface FileDropZoneProps {
  value: string                    // URL file yang sudah diupload
  onChange: (url: string) => void  // Callback saat file berhasil diupload
  label: string                    // Label field
  hint?: string                    // Teks bantuan (opsional)
  required?: boolean               // Tandai field sebagai required
  accept?: string                  // MIME types yang diterima (default: 'image/*')
  maxSizeMB?: number              // Ukuran file max dalam MB (default: 10)
  previewImageUrl?: string        // URL preview image
  onClear?: () => void            // Callback saat user clear file
}
```

**Fitur:**
- Drag & drop zone dengan visual feedback
- Click untuk membuka file picker
- Upload progress indicator
- File preview dengan tombol clear
- Error handling dengan toast notifications
- Format dan ukuran file constraints

**Contoh Penggunaan:**
```tsx
import { FileDropZone } from '@/widgets/FormPageTemplate'

export function MyForm() {
  const [form, setForm] = useState({
    logoUrl: '',
    certificateUrl: '',
  })

  const handleLogoChange = (url: string) => {
    setForm(prev => ({ ...prev, logoUrl: url }))
  }

  return (
    <div>
      <FileDropZone
        label="Logo Merek"
        hint="Opsional"
        value={form.logoUrl}
        onChange={handleLogoChange}
        accept="image/png,image/jpeg,image/webp"
        maxSizeMB={10}
        previewImageUrl={form.logoUrl}
        onClear={() => setForm(prev => ({ ...prev, logoUrl: '' }))}
      />

      <FileDropZone
        label="Sertifikat"
        hint="Format PDF max 5MB"
        value={form.certificateUrl}
        onChange={(url) => setForm(prev => ({ ...prev, certificateUrl: url }))}
        accept="application/pdf"
        maxSizeMB={5}
        previewImageUrl={form.certificateUrl}
      />
    </div>
  )
}
```

## Backend Integration

Kedua komponen menggunakan `mediaService` untuk upload file ke endpoint:

```
POST /api/v1/media/upload
```

Response format:
```json
{
  "url": "https://cdn.example.com/uploads/{company_id}/{domain}/{YYYY-MM}/{uuid}.{ext}",
  "file_id": "uuid-string"
}
```

## MIME Type Examples

**Images:**
```
image/png, image/jpeg, image/webp
```

**Documents:**
```
application/pdf, text/csv, 
application/vnd.openxmlformats-officedocument.spreadsheetml.sheet
```

**Single type:**
```
"image/*"           // Semua image formats
"application/pdf"   // Hanya PDF
```

## Styling & Themes

Komponen menggunakan CSS variables dari design system:
- `--color-primary` — Warna utama
- `--color-border` — Border color
- `--color-text-*` — Text colors
- `--space-*` — Spacing variables
- `--radius-md` — Border radius
- `--duration-fast` — Transition duration

## Architecture

Kedua komponen share:
- **mediaService** — Handle upload logic dengan XMLHttpRequest
- **Toast notifications** — User feedback
- **FormPageTemplate styles** — Consistent styling

Perbedaan utama:
- **FileUploadField** — Lebih simple, button-based
- **FileDropZone** — Lebih interactive, drag & drop + visual feedback

## Best Practices

1. **Validasi di form level** — Gunakan `required` prop untuk mandatory fields
2. **Clear callback** — Sediakan `onClear` untuk reset file
3. **Preview URL** — Pass `previewImageUrl` untuk show existing file
4. **File constraints** — Set `accept` dan `maxSizeMB` sesuai kebutuhan
5. **Error handling** — Toast notification sudah built-in, tidak perlu tambahan error handling

## Related Files

- `src/services/media.service.ts` — Media upload service
- `src/widgets/FormPageTemplate/FileUploadField.tsx` — Simple file input
- `src/widgets/FormPageTemplate/FileDropZone.tsx` — Drag & drop file input
