import { useCallback, useState } from 'react'
import { useQuery } from '@tanstack/react-query'
import { useDebounce } from './useDebounce'
import type { PaginatedResponse } from '@/types/api.types'
import type { ListParams } from '@/services/createEntityService'

interface SortState {
  key: string
  order: 'asc' | 'desc'
}

interface UseDataSourceConfig<T> {
  queryKey: string
  fetcher: (params: ListParams) => Promise<PaginatedResponse<T>>
  defaultPageSize?: number
  defaultSort?: SortState
}

export function useDataSource<T>({
  queryKey,
  fetcher,
  defaultPageSize = 25,
  defaultSort,
}: UseDataSourceConfig<T>) {
  const [page, setPageRaw] = useState(0)
  const [pageSize, setPageSizeRaw] = useState(defaultPageSize)
  const [sort, setSortRaw] = useState<SortState | undefined>(defaultSort)
  const [search, setSearchRaw] = useState('')
  const [filters, setFiltersRaw] = useState<Record<string, unknown>>({})

  const debouncedSearch = useDebounce(search, 300)

  const params: ListParams = {
    limit: pageSize,
    offset: page * pageSize,
    ...(sort ? { sort: sort.key, order: sort.order } : {}),
    ...(debouncedSearch ? { search: debouncedSearch } : {}),
    ...filters,
  }

  const { data, isLoading, error, refetch } = useQuery({
    queryKey: [queryKey, params],
    queryFn: () => fetcher(params),
    staleTime: 0,
    retry: 1,
  })

  const setPage = useCallback((p: number) => setPageRaw(p), [])

  const setPageSize = useCallback((size: number) => {
    setPageSizeRaw(size)
    setPageRaw(0)
  }, [])

  const setSort = useCallback((s: SortState | undefined) => {
    setSortRaw(s)
    setPageRaw(0)
  }, [])

  const setSearch = useCallback((s: string) => {
    setSearchRaw(s)
    setPageRaw(0)
  }, [])

  const setFilters = useCallback((f: Record<string, unknown>) => {
    setFiltersRaw(f)
    setPageRaw(0)
  }, [])

  return {
    data: data?.items ?? [],
    total: data?.total ?? 0,
    isLoading,
    error,
    refetch,
    pagination: { page, pageSize, total: data?.total ?? 0 },
    sort,
    search,
    filters,
    setPage,
    setPageSize,
    setSort,
    setSearch,
    setFilters,
  }
}
