import type { Dispatch, SetStateAction } from 'react'
import { Button, FormItem, Input, Modal } from '../../../../shared/components'
import type { BrowserCoreValidateResult } from '../../types'
import type { CoreEditForm } from '../coreManagement.types'

interface CoreEditModalProps {
  open: boolean
  isEditing: boolean
  form: CoreEditForm
  saving: boolean
  pathValidating: boolean
  pickingPath: boolean
  pathValidResult: BrowserCoreValidateResult | null
  setForm: Dispatch<SetStateAction<CoreEditForm>>
  onClose: () => void
  onSave: () => void
  onPickDirectory: () => void
}

export function CoreEditModal({
  open,
  isEditing,
  form,
  saving,
  pathValidating,
  pickingPath,
  pathValidResult,
  setForm,
  onClose,
  onSave,
  onPickDirectory,
}: CoreEditModalProps) {
  return (
    <Modal
      open={open}
      onClose={onClose}
      title={isEditing ? '编辑内核' : '新增内核'}
      width="500px"
      footer={
        <>
          <Button variant="secondary" onClick={onClose}>取消</Button>
          <Button onClick={onSave} loading={saving}>保存</Button>
        </>
      }
    >
      <div className="space-y-4">
        <FormItem label="内核名称" required>
          <Input
            value={form.coreName}
            onChange={e => setForm(prev => ({ ...prev, coreName: e.target.value }))}
            placeholder="例如：fingerprint-chromium 148"
          />
        </FormItem>
        <FormItem label="内核路径" required>
          {isEditing ? (
            <Input
              value={form.corePath}
              onChange={e => setForm(prev => ({ ...prev, corePath: e.target.value }))}
              placeholder="内核安装目录的绝对路径"
            />
          ) : (
            <div className="flex gap-2">
              <Input
                value={form.corePath}
                readOnly
                placeholder="选择内核目录或 macOS 的 .app"
                className="flex-1"
              />
              <Button variant="secondary" onClick={onPickDirectory} loading={pickingPath}>
                选择目录
              </Button>
            </div>
          )}
          {pathValidating && (
            <p className="text-xs text-[var(--color-text-muted)] mt-1">验证中...</p>
          )}
          {!pathValidating && pathValidResult && (
            <p className={`text-xs mt-1 ${pathValidResult.valid ? 'text-green-600' : 'text-red-500'}`}>
              {pathValidResult.message}
            </p>
          )}
        </FormItem>
      </div>
    </Modal>
  )
}
