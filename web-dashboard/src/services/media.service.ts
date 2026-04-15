import { useAuthStore } from '@/stores/auth.store'

const BASE_URL = import.meta.env.VITE_API_BASE_URL ?? 'http://localhost:8080'

interface UploadResponse {
  url: string
  file_id: string
}

/**
 * Media service untuk menangani file upload ke endpoint /api/v1/media/upload
 * Menggunakan FormData untuk mengirim file binari
 */
export const mediaService = {
  /**
   * Upload file ke media storage
   * @param file File object dari file input
   * @param onProgress Optional callback untuk tracking progress (bytes uploaded)
   * @returns Promise dengan URL dan file_id
   */
  async uploadFile(file: File, onProgress?: (progress: number) => void): Promise<UploadResponse> {
    const authState = useAuthStore.getState()
    const formData = new FormData()
    formData.append('file', file)

    const headers: Record<string, string> = {}
    if (authState.token) {
      headers['Authorization'] = `Bearer ${authState.token}`
    }
    if (authState.selectedGroup) {
      headers['X-Company-Group-ID'] = authState.selectedGroup.id
    }
    if (authState.selectedCompany) {
      headers['X-Company-ID'] = authState.selectedCompany.id
    }

    return new Promise((resolve, reject) => {
      const xhr = new XMLHttpRequest()

      // Track upload progress
      if (onProgress) {
        xhr.upload.addEventListener('progress', (event) => {
          if (event.lengthComputable) {
            const progress = Math.round((event.loaded / event.total) * 100)
            onProgress(progress)
          }
        })
      }

      xhr.addEventListener('load', () => {
        if (xhr.status === 401) {
          authState.logout()
          reject(new Error('Sesi telah berakhir, silakan login kembali'))
          return
        }

        if (xhr.status >= 400) {
          try {
            const data = JSON.parse(xhr.responseText) as Record<string, unknown>
            const message = typeof data['error'] === 'string' ? data['error'] : 'Gagal upload file'
            reject(new Error(message))
          } catch {
            reject(new Error('Gagal upload file'))
          }
          return
        }

        try {
          const response = JSON.parse(xhr.responseText) as UploadResponse
          resolve(response)
        } catch {
          reject(new Error('Response format tidak valid'))
        }
      })

      xhr.addEventListener('error', () => {
        reject(new Error('Terjadi kesalahan saat upload'))
      })

      xhr.addEventListener('abort', () => {
        reject(new Error('Upload dibatalkan'))
      })

      xhr.open('POST', `${BASE_URL}/api/v1/media/upload`)

      // Set headers
      Object.entries(headers).forEach(([key, value]) => {
        xhr.setRequestHeader(key, value)
      })

      xhr.send(formData)
    })
  },
}
