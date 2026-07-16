import { Link } from 'react-router-dom'
import { XCircle } from 'lucide-react'
import { Button, Modal } from '../../../../shared/components'
import { BrowserProfileCopyForm } from '../../components/BrowserProfileCopyForm'
import { KeywordsModal } from '../../components/KeywordsModal'
import type { BrowserProfile, BrowserProfileCopyOptions } from '../../types'

interface BrowserListDialogsProps {
  proxyErrorModal: boolean
  pendingStartId: string | null
  proxyErrorMsg: string
  onCloseProxyError: () => void
  onStartDirect: () => void
  startingDirect: boolean
  kwModal: { open: boolean; profile: BrowserProfile | null }
  onCloseKeywords: () => void
  onKeywordsSaved: (keywords: string[]) => void
  copyModal: { open: boolean; profile: BrowserProfile | null }
  copyName: string
  copyOptions: BrowserProfileCopyOptions
  onCopyNameChange: (value: string) => void
  onCopyOptionsChange: (value: BrowserProfileCopyOptions) => void
  onCloseCopy: () => void
  onConfirmCopy: () => void
  copyConfirmDisabled: boolean
  copying: boolean
  deleteConfirm: { open: boolean; mode: 'single' | 'batch'; profileName?: string; count: number }
  deleting: boolean
  onCloseDeleteConfirm: () => void
  onConfirmDelete: () => void
  opError: string
  onCloseOpError: () => void
}

export function BrowserListDialogs({
  proxyErrorModal,
  pendingStartId,
  proxyErrorMsg,
  onCloseProxyError,
  onStartDirect,
  startingDirect,
  kwModal,
  onCloseKeywords,
  onKeywordsSaved,
  copyModal,
  copyName,
  copyOptions,
  onCopyNameChange,
  onCopyOptionsChange,
  onCloseCopy,
  onConfirmCopy,
  copyConfirmDisabled,
  copying,
  deleteConfirm,
  deleting,
  onCloseDeleteConfirm,
  onConfirmDelete,
  opError,
  onCloseOpError,
}: BrowserListDialogsProps) {
  return (
    <>
      <Modal
        open={proxyErrorModal}
        onClose={onCloseProxyError}
        title="代理链路不可用"
        width="420px"
        footer={
          <>
            <Button variant="secondary" onClick={onCloseProxyError} disabled={startingDirect}>取消</Button>
            {pendingStartId && (
              <Button variant="secondary" onClick={onStartDirect} loading={startingDirect}>
                直连启动
              </Button>
            )}
            {pendingStartId && (
              <Link to={`/browser/edit/${pendingStartId}`}>
                <Button onClick={onCloseProxyError} disabled={startingDirect}>去修改代理</Button>
              </Link>
            )}
          </>
        }
      >
        <div className="space-y-3">
          <div className="flex items-start gap-3 p-3 rounded-lg bg-[var(--color-bg-secondary)]">
            <XCircle className="w-5 h-5 text-red-500 mt-0.5 shrink-0" />
            <p className="text-sm text-[var(--color-text-primary)]">{proxyErrorMsg}</p>
          </div>
          <p className="text-sm text-[var(--color-text-muted)]">请前往编辑页面重新选择可用链路；如果是订阅导入，先刷新订阅并确认该节点仍存在。</p>
        </div>
      </Modal>

      {kwModal.profile && (
        <KeywordsModal
          open={kwModal.open}
          profileId={kwModal.profile.profileId}
          profileName={kwModal.profile.profileName}
          initialKeywords={kwModal.profile.keywords || []}
          onClose={onCloseKeywords}
          onSaved={onKeywordsSaved}
        />
      )}

      <Modal
        open={copyModal.open}
        onClose={onCloseCopy}
        title="复制实例"
        width="720px"
        footer={
          <>
            <Button variant="secondary" onClick={onCloseCopy}>取消</Button>
            <Button onClick={onConfirmCopy} loading={copying} disabled={copyConfirmDisabled}>确认复制</Button>
          </>
        }
      >
        <BrowserProfileCopyForm
          sourceName={copyModal.profile?.profileName}
          copyName={copyName}
          copyOptions={copyOptions}
          onCopyNameChange={onCopyNameChange}
          onCopyOptionsChange={onCopyOptionsChange}
          autoFocusName
        />
      </Modal>

      <Modal
        open={deleteConfirm.open}
        onClose={onCloseDeleteConfirm}
        title="删除实例"
        width="420px"
        footer={
          <>
            <Button variant="secondary" onClick={onCloseDeleteConfirm} disabled={deleting}>取消</Button>
            <Button variant="danger" onClick={onConfirmDelete} loading={deleting}>确定删除</Button>
          </>
        }
      >
        <div className="space-y-2 text-sm text-[var(--color-text-secondary)]">
          <p>
            {deleteConfirm.mode === 'batch'
              ? `确定删除选中的 ${deleteConfirm.count} 个实例？`
              : `确定删除实例「${deleteConfirm.profileName || '未命名实例'}」？`}
          </p>
          <p className="text-red-500">这会删除配置、浏览器用户数据、快照、快捷码和插件绑定，删除后不可恢复。</p>
        </div>
      </Modal>

      <Modal
        open={!!opError}
        onClose={onCloseOpError}
        title="操作失败"
        width="420px"
        footer={<Button onClick={onCloseOpError}>知道了</Button>}
      >
        <div className="text-[var(--color-text-secondary)] whitespace-pre-line">{opError}</div>
      </Modal>
    </>
  )
}
