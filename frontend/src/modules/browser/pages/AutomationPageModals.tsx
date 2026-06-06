import { Button, FormItem, Input, Modal, Select, Textarea } from "../../../shared/components";
import { AUTOMATION_SCRIPT_TYPE_OPTIONS, type AutomationScriptType } from "../automationScripts";
import type { ImportMode } from "./AutomationPage.helpers";

interface CreateAutomationScriptModalProps {
  open: boolean;
  busyAction: "none" | "create" | "import";
  createName: string;
  createType: AutomationScriptType;
  onClose: () => void;
  onCreate: () => Promise<void>;
  onCreateNameChange: (value: string) => void;
  onCreateTypeChange: (value: AutomationScriptType) => void;
}

export function CreateAutomationScriptModal({
  open,
  busyAction,
  createName,
  createType,
  onClose,
  onCreate,
  onCreateNameChange,
  onCreateTypeChange,
}: CreateAutomationScriptModalProps) {
  return (
      <Modal
        open={open}
        onClose={onClose}
        title="新建脚本"
        width="460px"
        footer={
          <>
            <Button
              variant="secondary"
              onClick={onClose}
              disabled={busyAction !== "none"}
            >
              取消
            </Button>
            <Button
              onClick={() => void onCreate()}
              loading={busyAction === "create"}
            >
              创建
            </Button>
          </>
        }
      >
        <div className="space-y-4">
          <FormItem label="脚本名称">
            <Input
              value={createName}
              onChange={(event) => onCreateNameChange(event.target.value)}
              placeholder="例如：接管页面并截图"
            />
          </FormItem>
          <FormItem label="脚本类型">
            <Select
              value={createType}
              options={AUTOMATION_SCRIPT_TYPE_OPTIONS}
              onChange={(event) =>
                onCreateTypeChange(event.target.value as AutomationScriptType)
              }
            />
          </FormItem>
        </div>
      </Modal>
  );
}

interface ImportAutomationScriptModalProps {
  open: boolean;
  busyAction: "none" | "create" | "import";
  importMode: ImportMode;
  importText: string;
  remoteURL: string;
  gitURL: string;
  gitRef: string;
  gitScriptPath: string;
  onClose: () => void;
  onImport: () => Promise<void>;
  onImportModeChange: (value: ImportMode) => void;
  onImportTextChange: (value: string) => void;
  onRemoteURLChange: (value: string) => void;
  onGitURLChange: (value: string) => void;
  onGitRefChange: (value: string) => void;
  onGitScriptPathChange: (value: string) => void;
}

export function ImportAutomationScriptModal({
  open,
  busyAction,
  importMode,
  importText,
  remoteURL,
  gitURL,
  gitRef,
  gitScriptPath,
  onClose,
  onImport,
  onImportModeChange,
  onImportTextChange,
  onRemoteURLChange,
  onGitURLChange,
  onGitRefChange,
  onGitScriptPathChange,
}: ImportAutomationScriptModalProps) {
  return (
      <Modal
        open={open}
        onClose={onClose}
        title="导入脚本"
        width="720px"
        footer={
          <>
            <Button
              variant="secondary"
              onClick={onClose}
              disabled={busyAction !== "none"}
            >
              取消
            </Button>
            <Button
              onClick={() => void onImport()}
              loading={busyAction === "import"}
            >
              导入
            </Button>
          </>
        }
      >
        <div className="space-y-4">
          <div className="flex flex-wrap gap-2">
            {[
              { value: "text", label: "文本" },
              { value: "local-file", label: "本地文件" },
              { value: "local-dir", label: "本地目录" },
              { value: "local-library", label: "脚本库" },
              { value: "remote-url", label: "远程 URL" },
              { value: "git", label: "Git" },
            ].map((item) => (
              <Button
                key={item.value}
                size="sm"
                variant={importMode === item.value ? "primary" : "secondary"}
                onClick={() => onImportModeChange(item.value as ImportMode)}
                disabled={busyAction !== "none"}
              >
                {item.label}
              </Button>
            ))}
          </div>

          {importMode === "text" ? (
            <>
              <div className="text-sm text-[var(--color-text-secondary)]">
                支持导入导出的脚本 JSON，导入后会按草稿保存。
              </div>
              <FormItem label="脚本 JSON">
                <Textarea
                  rows={18}
                  value={importText}
                  onChange={(event) => onImportTextChange(event.target.value)}
                  className="font-mono"
                  placeholder='{"manifest":{"name":"示例脚本"}}'
                />
              </FormItem>
            </>
          ) : null}

          {importMode === "local-file" ? (
            <div className="rounded-xl border border-[var(--color-border-default)] bg-[var(--color-bg-secondary)] px-4 py-4 text-sm text-[var(--color-text-secondary)]">
              导入时会弹出文件选择框。支持单个 `.js/.cjs/.mjs` 脚本文件、导出的
              `.json` 模板，或标准 `.zip` 脚本包。`.ts/.cts/.mts` 仅在设置页开启 TypeScript 导入构建后支持。
            </div>
          ) : null}

          {importMode === "local-dir" ? (
            <div className="rounded-xl border border-[var(--color-border-default)] bg-[var(--color-bg-secondary)] px-4 py-4 text-sm text-[var(--color-text-secondary)]">
              导入时会弹出目录选择框。适合导入一整套本地脚本目录，或 Git 拉下来的脚本包目录。目录里的 `.ts/.cts/.mts` 入口也需要先在设置页开启 TypeScript 导入构建。
            </div>
          ) : null}

          {importMode === "local-library" ? (
            <div className="rounded-xl border border-[var(--color-border-default)] bg-[var(--color-bg-secondary)] px-4 py-4 text-sm text-[var(--color-text-secondary)]">
              导入时会弹出目录选择框。系统会扫描所选目录下的脚本包并批量导入，来源按本地目录记录，后续刷新不走 Git。
            </div>
          ) : null}

          {importMode === "remote-url" ? (
            <div className="space-y-4">
              <div className="text-sm text-[var(--color-text-secondary)]">
                适合导入单个远程脚本文件、导出的脚本 JSON，或标准脚本 ZIP。多文件仓库也可以继续使用 Git 导入；远程 `.ts/.cts/.mts` 同样要求设置页已开启 TypeScript 导入构建。
              </div>
              <FormItem label="远程地址">
                <Input
                  value={remoteURL}
                  onChange={(event) => onRemoteURLChange(event.target.value)}
                  placeholder="https://example.com/script.cjs"
                />
              </FormItem>
            </div>
          ) : null}

          {importMode === "git" ? (
            <div className="space-y-4">
              <div className="text-sm text-[var(--color-text-secondary)]">
                会先拉取仓库，再把脚本快照导入当前项目。可以只填一个脚本子目录，系统只扫描那个目录；不填时才会按仓库根目录解析。若入口是 `.ts/.cts/.mts`，需要设置页已开启 TypeScript 导入构建。
              </div>
              <FormItem label="仓库地址">
                <Input
                  value={gitURL}
                  onChange={(event) => onGitURLChange(event.target.value)}
                  placeholder="https://github.com/example/automation-scripts.git"
                />
              </FormItem>
              <div className="grid grid-cols-1 gap-4 md:grid-cols-2">
                <FormItem label="分支 / Tag / Commit">
                  <Input
                    value={gitRef}
                    onChange={(event) => onGitRefChange(event.target.value)}
                    placeholder="main"
                  />
                </FormItem>
                <FormItem label="脚本路径">
                  <Input
                    value={gitScriptPath}
                    onChange={(event) => onGitScriptPathChange(event.target.value)}
                    placeholder="scripts/demo（留空=仓库根目录）"
                  />
                </FormItem>
              </div>
            </div>
          ) : null}
        </div>
      </Modal>
  );
}
