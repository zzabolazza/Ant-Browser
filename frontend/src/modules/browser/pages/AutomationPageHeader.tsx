import { History, PlusSquare, RefreshCw, Upload, Wrench } from "lucide-react";
import { Button } from "../../../shared/components";

interface AutomationPageHeaderProps {
  refreshing: boolean;
  onRefresh: () => void;
  onCreate: () => void;
  onImport: () => void;
  onOpenHistory: () => void;
  onOpenToolbox: () => void;
}

export function AutomationPageHeader({
  refreshing,
  onRefresh,
  onCreate,
  onImport,
  onOpenHistory,
  onOpenToolbox,
}: AutomationPageHeaderProps) {
  return (
      <div className="flex items-center justify-between gap-3 flex-wrap">
        <h1 className="text-xl font-semibold text-[var(--color-text-primary)]">
          脚本管理
        </h1>
        <div className="flex flex-wrap gap-2">
          <Button
            size="sm"
            variant="secondary"
            onClick={onRefresh}
            loading={refreshing}
          >
            <RefreshCw className="h-4 w-4" />
            刷新
          </Button>
          <Button size="sm" onClick={onCreate}>
            <PlusSquare className="h-4 w-4" />
            新建脚本
          </Button>
          <Button
            size="sm"
            variant="secondary"
            onClick={onImport}
          >
            <Upload className="h-4 w-4" />
            导入脚本
          </Button>
          <Button
            size="sm"
            variant="secondary"
            onClick={onOpenHistory}
          >
            <History className="h-4 w-4" />
            调用记录
          </Button>
          <Button
            size="sm"
            variant="secondary"
            onClick={onOpenToolbox}
          >
            <Wrench className="h-4 w-4" />
            工具箱
          </Button>
        </div>
      </div>

  );
}
