import { createContext, useContext } from 'react'
import type { FlowStep } from '@/widgets/FlowWidget/FlowWidget'

export interface FlowInfoContextValue {
  openFlowInfo: (step: FlowStep, accentColor: string) => void
  closeFlowInfo: () => void
  activeInfo: { step: FlowStep; accentColor: string } | null
}

export const FlowInfoContext = createContext<FlowInfoContextValue | null>(null)

export function useFlowInfo(): FlowInfoContextValue {
  const ctx = useContext(FlowInfoContext)
  if (!ctx) throw new Error('useFlowInfo must be used inside FlowInfoProvider')
  return ctx
}
