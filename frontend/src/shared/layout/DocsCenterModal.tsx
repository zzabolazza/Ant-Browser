import { Suspense, lazy } from 'react'
import { Modal } from '../components'

const LaunchDocsPanel = lazy(() =>
  import('../../modules/browser/pages/launchApiDocs/LaunchDocsPanel').then((module) => ({
    default: module.LaunchDocsPanel,
  }))
)

interface DocsCenterModalProps {
  open: boolean
  onClose: () => void
}

export function DocsCenterModal({ open, onClose }: DocsCenterModalProps) {
  return (
    <Modal
      open={open}
      onClose={onClose}
      title="文档中心"
      width="1200px"
      padding={false}
    >
      <div className="h-[72vh] min-h-[480px]">
        <Suspense fallback={(
          <div className="flex h-full items-center justify-center text-sm text-[var(--color-text-muted)]">
            加载文档中...
          </div>
        )}>
          <LaunchDocsPanel embedded />
        </Suspense>
      </div>
    </Modal>
  )
}
