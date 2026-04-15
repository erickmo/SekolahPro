import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { ToastProvider } from '@/widgets/Toast/Toast'

const queryClient = new QueryClient({
  defaultOptions: {
    queries: {
      staleTime: 30_000,
      retry: 1,
      refetchOnWindowFocus: false,
    },
  },
})

interface ProvidersProps {
  children: React.ReactNode
}

export function Providers({ children }: ProvidersProps) {
  return (
    <QueryClientProvider client={queryClient}>
      {children}
      <ToastProvider />
    </QueryClientProvider>
  )
}
