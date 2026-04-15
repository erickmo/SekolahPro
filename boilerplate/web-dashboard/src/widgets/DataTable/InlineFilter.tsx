import { useState } from 'react'
import { Plus } from 'lucide-react'
import type { ActiveFilter, FilterDef, FilterOperator } from './filter.types'
import { getOperatorsForDef, OPERATOR_LABELS } from './filter.utils'
import { FilterValueInput } from './FilterValueInput'
import styles from './InlineFilter.module.css'

interface InlineFilterProps {
  defs: FilterDef[]
  onApply: (filter: ActiveFilter) => void
}

export function InlineFilter({ defs, onApply }: InlineFilterProps) {
  const firstDef = defs[0]
  const firstOps = firstDef ? getOperatorsForDef(firstDef) : (['eq'] as FilterOperator[])

  const [key, setKey] = useState(firstDef?.key ?? '')
  const [op, setOp] = useState<FilterOperator>(firstOps[0])
  const [value, setValue] = useState<unknown>('')

  if (!firstDef) return null

  const def = defs.find((d) => d.key === key) ?? firstDef
  const operators = getOperatorsForDef(def)

  function handleFieldChange(newKey: string) {
    const newDef = defs.find((d) => d.key === newKey)
    if (!newDef) return
    const ops = getOperatorsForDef(newDef)
    const defaultVal = newDef.type === 'date-range' || ops[0] === 'between' ? ['', ''] : ''
    setKey(newKey)
    setOp(ops[0])
    setValue(defaultVal)
  }

  function handleOpChange(newOp: FilterOperator) {
    const isBetween = newOp === 'between'
    const wasBetween = op === 'between'
    setValue(isBetween ? ['', ''] : wasBetween ? '' : value)
    setOp(newOp)
  }

  function handleApply() {
    const v = value
    const isEmpty =
      v === '' ||
      v === null ||
      v === undefined ||
      (Array.isArray(v) && (v as string[]).every((x) => x === ''))
    if (isEmpty) return
    const id = Math.random().toString(36).slice(2)
    onApply({ id, key, operator: op, value: v })
    // Reset value after applying
    const ops = getOperatorsForDef(def)
    setValue(def.type === 'date-range' || ops[0] === 'between' ? ['', ''] : '')
  }

  return (
    <div className={styles.wrap}>
      <span className={styles.label}>Filter cepat</span>
      <select
        className={styles.fieldSelect}
        value={key}
        onChange={(e) => handleFieldChange(e.target.value)}
      >
        {defs.map((d) => (
          <option key={d.key} value={d.key}>
            {d.label}
          </option>
        ))}
      </select>
      <select
        className={styles.opSelect}
        value={op}
        onChange={(e) => handleOpChange(e.target.value as FilterOperator)}
      >
        {operators.map((o) => (
          <option key={o} value={o}>
            {OPERATOR_LABELS[o]}
          </option>
        ))}
      </select>
      <div className={styles.valueWrap}>
        <FilterValueInput def={def} operator={op} value={value} onValueChange={setValue} />
      </div>
      <button className={styles.applyBtn} onClick={handleApply} title="Terapkan filter">
        <Plus size={13} />
        <span>Terapkan</span>
      </button>
    </div>
  )
}
