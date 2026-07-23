import type { Dispatch, SetStateAction } from 'react'
import { Edit2, Save, Settings, X } from 'lucide-react'
import { Button, Card, FormItem, Input, Switch, Textarea } from '../../../../shared/components'
import type { BrowserSettings } from '../../types'
import type { CoreSettingsForm } from '../coreManagement.types'

interface CoreSettingsCardProps {
  settings: BrowserSettings
  form: CoreSettingsForm
  editing: boolean
  saving: boolean
  setForm: Dispatch<SetStateAction<CoreSettingsForm>>
  onEdit: () => void
  onCancel: () => void
  onSave: () => void
}

const settingsValueClass = 'min-h-10 overflow-auto rounded-lg border border-[var(--color-border-default)] bg-[var(--color-bg-muted)] px-3 py-2 text-[12.5px] leading-5 text-[var(--color-text-primary)]'

export function CoreSettingsCard({
  settings,
  form,
  editing,
  saving,
  setForm,
  onEdit,
  onCancel,
  onSave,
}: CoreSettingsCardProps) {
  return (
    <Card padding="sm">
      <div className="mb-4 flex items-center justify-between">
        <div className="flex items-center gap-2">
          <Settings className="h-4 w-4 text-[var(--color-text-muted)]" />
          <h3 className="text-[13.5px] font-bold text-[var(--color-text-primary)]">全局启动设置</h3>
        </div>
        {editing ? (
          <div className="flex items-center gap-2">
            <Button size="sm" variant="ghost" onClick={onCancel} disabled={saving}>
              <X className="h-4 w-4" />
              取消
            </Button>
            <Button size="sm" onClick={onSave} loading={saving}>
              <Save className="h-4 w-4" />
              保存
            </Button>
          </div>
        ) : (
          <Button size="sm" variant="ghost" onClick={onEdit}>
            <Edit2 className="h-4 w-4" />
            编辑
          </Button>
        )}
      </div>
      {editing ? (
        <InlineSettingsForm form={form} setForm={setForm} />
      ) : (
        <div className="grid grid-cols-1 gap-4 md:grid-cols-2">
          <SettingsList label="默认启动参数" values={settings.defaultLaunchArgs} mono />
          <SettingsList label="默认打开页面" values={settings.defaultStartUrls} />
          <SettingsList label="默认指纹参数" values={settings.defaultFingerprintArgs} mono />
          <SettingsValue label="用户数据根目录" value={settings.userDataRoot || '-'} mono />
          <SettingsValue label="恢复上次标签页" value={settings.restoreLastSession ? '开启' : '关闭'} />
          <SettingsValue label="轻启动模式" value={settings.lightStartEnabled ? '开启' : '关闭'} />
          <SettingsValue label="启动就绪超时" value={`${settings.startReadyTimeoutMs} ms`} />
          <SettingsValue label="启动稳定窗口" value={`${settings.startStableWindowMs} ms`} />
        </div>
      )}
    </Card>
  )
}

function InlineSettingsForm({
  form,
  setForm,
}: {
  form: CoreSettingsForm
  setForm: Dispatch<SetStateAction<CoreSettingsForm>>
}) {
  return (
    <div className="grid grid-cols-1 gap-4 md:grid-cols-2">
      <FormItem label="默认启动参数" hint="每行一个参数">
        <Textarea
          value={form.defaultLaunchArgs}
          onChange={e => setForm(prev => ({ ...prev, defaultLaunchArgs: e.target.value }))}
          rows={4}
          placeholder="例如：--disable-sync"
          className="font-mono text-[12.5px]"
        />
      </FormItem>
      <FormItem label="默认打开页面" hint="每行一个 URL">
        <Textarea
          value={form.defaultStartUrls}
          onChange={e => setForm(prev => ({ ...prev, defaultStartUrls: e.target.value }))}
          rows={4}
          placeholder="例如：https://example.com"
        />
      </FormItem>
      <FormItem label="默认指纹参数" hint="每行一个参数">
        <Textarea
          value={form.defaultFingerprintArgs}
          onChange={e => setForm(prev => ({ ...prev, defaultFingerprintArgs: e.target.value }))}
          rows={4}
          placeholder={"例如：--fingerprint-brand=Chrome\n--fingerprint-platform=macos"}
          className="font-mono text-[12.5px]"
        />
      </FormItem>
      <FormItem label="用户数据根目录" hint="绝对路径，默认位于用户状态目录下的 data">
        <Input
          value={form.userDataRoot}
          onChange={e => setForm(prev => ({ ...prev, userDataRoot: e.target.value }))}
          placeholder="例如：/Users/you/Library/Application Support/facade/data"
          className="font-mono text-[12.5px]"
        />
      </FormItem>
      <ToggleField
        label="恢复上次标签页"
        description="下次启动时恢复之前的标签页和窗口"
        checked={form.restoreLastSession}
        onChange={checked => setForm(prev => ({ ...prev, restoreLastSession: checked }))}
      />
      <ToggleField
        label="轻启动模式"
        description="先启动空白页，就绪后再打开默认页面"
        checked={form.lightStartEnabled}
        onChange={checked => setForm(prev => ({ ...prev, lightStartEnabled: checked }))}
      />
      <FormItem label="启动就绪超时（毫秒）" hint="最小 1000">
        <Input
          type="number"
          min={1000}
          step={500}
          value={form.startReadyTimeoutMs}
          onChange={e => setForm(prev => ({
            ...prev,
            startReadyTimeoutMs: Math.max(1000, Number(e.target.value) || 3000),
          }))}
        />
      </FormItem>
      <FormItem label="启动稳定窗口（毫秒）" hint="建议 1200-3000">
        <Input
          type="number"
          min={0}
          step={100}
          value={form.startStableWindowMs}
          onChange={e => setForm(prev => ({
            ...prev,
            startStableWindowMs: Math.max(0, Number(e.target.value) || 1200),
          }))}
        />
      </FormItem>
    </div>
  )
}

function ToggleField({
  label,
  description,
  checked,
  onChange,
}: {
  label: string
  description: string
  checked: boolean
  onChange: (checked: boolean) => void
}) {
  return (
    <div className="flex min-h-[68px] items-center justify-between rounded-lg border border-[var(--color-border-default)] px-3 py-2">
      <div>
        <div className="text-[13px] font-medium text-[var(--color-text-primary)]">{label}</div>
        <div className="mt-0.5 text-[11.5px] text-[var(--color-text-muted)]">{description}</div>
      </div>
      <Switch checked={checked} onChange={onChange} />
    </div>
  )
}

function SettingsValue({ label, value, mono = false }: { label: string; value: string; mono?: boolean }) {
  return (
    <div>
      <p className="mb-1.5 text-[11.5px] font-semibold text-[var(--color-text-muted)]">{label}</p>
      <div className={`${settingsValueClass} break-all ${mono ? 'font-mono' : ''}`}>
        {value}
      </div>
    </div>
  )
}

function SettingsList({ label, values, mono = false }: { label: string; values: string[]; mono?: boolean }) {
  return (
    <div>
      <p className="mb-1.5 text-[11.5px] font-semibold text-[var(--color-text-muted)]">{label}</p>
      {values.length > 0 ? (
        <pre className={`${settingsValueClass} ${mono ? 'font-mono' : ''}`}>
          {values.join('\n')}
        </pre>
      ) : (
        <div className={settingsValueClass}>
          -
        </div>
      )}
    </div>
  )
}
