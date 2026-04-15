import { useState, useCallback } from 'react'
import { FlowInfoContext } from './useFlowInfo'
import type { FlowStep } from '@/widgets/FlowWidget/FlowWidget'

export function FlowInfoProvider({ children }: { children: React.ReactNode }) {
  const [activeInfo, setActiveInfo] = useState<{ step: FlowStep; accentColor: string } | null>(null)

  const openFlowInfo = useCallback((step: FlowStep, accentColor: string) => {
    setActiveInfo({ step, accentColor })
  }, [])

  const closeFlowInfo = useCallback(() => setActiveInfo(null), [])

  return (
    <FlowInfoContext.Provider value={{ openFlowInfo, closeFlowInfo, activeInfo }}>
      {children}
    </FlowInfoContext.Provider>
  )
}
