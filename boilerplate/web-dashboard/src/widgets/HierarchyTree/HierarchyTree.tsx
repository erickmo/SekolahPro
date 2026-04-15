import { useState } from 'react'
import { ChevronRight, ChevronDown } from 'lucide-react'
import styles from './HierarchyTree.module.css'

export interface HierarchyNode {
  id: string
  label: string
  [key: string]: any
  children?: HierarchyNode[]
}

export interface HierarchyTreeProps<T extends HierarchyNode> {
  nodes: T[]
  onNodeClick: (node: T) => void
  renderNode?: (node: T) => React.ReactNode
  onLoadChildren?: (nodeId: string) => Promise<void>
  isLoading?: boolean
  emptyMessage?: string
  maxHeight?: string
}

function TreeNodeItem<T extends HierarchyNode>({
  node,
  onNodeClick,
  renderNode,
  onLoadChildren,
  depth = 0,
}: {
  node: T
  onNodeClick: (node: T) => void
  renderNode?: (node: T) => React.ReactNode
  onLoadChildren?: (nodeId: string) => Promise<void>
  depth?: number
}) {
  const [expanded, setExpanded] = useState(false)
  const [loading, setLoading] = useState(false)
  const hasChildren = node.children && node.children.length > 0

  const handleToggle = async (e: React.MouseEvent) => {
    e.stopPropagation()
    if (!expanded && !hasChildren && onLoadChildren) {
      setLoading(true)
      await onLoadChildren(node.id)
      setLoading(false)
    }
    setExpanded(!expanded)
  }

  return (
    <div>
      <div
        className={styles.nodeRow}
        style={{ paddingLeft: `${depth * 1.5}rem` }}
        onClick={() => onNodeClick(node)}
      >
        <button
          className={styles.expandBtn}
          onClick={handleToggle}
          disabled={loading || (!hasChildren && !onLoadChildren)}
        >
          {expanded ? (
            <ChevronDown size={16} />
          ) : (
            <ChevronRight size={16} />
          )}
        </button>
        <div className={styles.nodeContent}>
          {renderNode ? renderNode(node) : <span>{node.label}</span>}
        </div>
      </div>

      {expanded && hasChildren && (
        <div>
          {node.children?.map((child) => (
            <TreeNodeItem
              key={child.id}
              node={child as T}
              onNodeClick={onNodeClick}
              renderNode={renderNode}
              onLoadChildren={onLoadChildren}
              depth={depth + 1}
            />
          ))}
        </div>
      )}
    </div>
  )
}

export function HierarchyTree<T extends HierarchyNode>({
  nodes,
  onNodeClick,
  renderNode,
  onLoadChildren,
  isLoading = false,
  emptyMessage = 'No items',
  maxHeight = 'calc(100vh - 300px)',
}: HierarchyTreeProps<T>) {
  return (
    <div className={styles.container} style={{ maxHeight }}>
      {isLoading ? (
        <div className={styles.loading}>Loading...</div>
      ) : nodes.length === 0 ? (
        <div className={styles.empty}>{emptyMessage}</div>
      ) : (
        <div className={styles.tree}>
          {nodes.map((node) => (
            <TreeNodeItem
              key={node.id}
              node={node}
              onNodeClick={onNodeClick}
              renderNode={renderNode}
              onLoadChildren={onLoadChildren}
            />
          ))}
        </div>
      )}
    </div>
  )
}
