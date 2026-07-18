import { Link, useLocation } from "react-router-dom";
import {
  Activity,
  Bookmark,
  BookOpen,
  FileText,
  LayoutDashboard,
  ListChecks,
  Monitor,
  Settings,
  Database,
  Layers,
  PieChart,
  Cpu,
  Globe,
  Bot,
  Puzzle,
  Tag,
  ChevronLeft,
  ChevronRight,
  type LucideIcon,
} from "lucide-react";
import clsx from "clsx";
import { useLayoutStore } from "../../store/layoutStore";
import { projectConfig, navigationConfig } from "../../config";

const iconMap: Record<string, LucideIcon> = {
  LayoutDashboard,
  Settings,
  Database,
  Layers,
  PieChart,
  Monitor,
  ListChecks,
  Activity,
  FileText,
  Cpu,
  Globe,
  Bot,
  Puzzle,
  Bookmark,
  BookOpen,
  Tag,
};

function getIcon(iconName: string): LucideIcon {
  return iconMap[iconName] || LayoutDashboard;
}

export function Sidebar() {
  const location = useLocation();
  const { sidebarCollapsed, toggleSidebar } = useLayoutStore();

  return (
    <aside
      className={clsx(
        "flex shrink-0 flex-col bg-gradient-to-b from-[#12141c] to-[#181b26] text-[#c9cbdb] transition-[width] duration-200",
        sidebarCollapsed ? "w-16 px-2" : "w-56 px-3 max-sm:w-16 max-sm:px-2",
      )}
    >
      <div
        className={clsx(
          "flex h-[66px] items-center gap-2.5",
          sidebarCollapsed ? "justify-center" : "px-2 max-sm:justify-center max-sm:px-0",
        )}
      >
        <img
          src="/appicon.png"
          alt="Facade"
          className="h-[30px] w-[30px] shrink-0 rounded-lg object-cover shadow-[inset_0_0_0_1px_rgb(255_255_255_/_0.14)]"
        />
        {!sidebarCollapsed && (
          <div className="min-w-0 text-center">
            <div className="truncate text-[15px] font-extrabold tracking-tight text-white max-sm:hidden">
              {projectConfig.name}
            </div>
            <div className="font-mono text-[10px] tracking-[0.08em] text-[#6d7186] max-sm:hidden">
              Endless Faces
            </div>
          </div>
        )}
      </div>

      <nav className="flex-1 overflow-y-auto py-2">
        <div className="space-y-1">
          {navigationConfig.map((item, index) => {
            const Icon = getIcon(item.icon);
            const isActive =
              location.pathname === item.path ||
              (item.path !== "/" &&
                location.pathname.startsWith(`${item.path}/`));
            const label = String(index + 1).padStart(2, "0");

            return (
              <Link
                key={item.path}
                to={item.path}
                title={item.name}
                aria-label={item.name}
                className={clsx(
                  "group relative flex items-center rounded-lg border transition-colors duration-150",
                  isActive
                    ? "border-white/10 bg-[#1f2333] text-white"
                    : "border-transparent text-[#a9acc0] hover:bg-white/[0.06] hover:text-[#e7e8f2]",
                  sidebarCollapsed
                    ? "mx-auto h-10 w-10 justify-center"
                    : "gap-3 px-3 py-2.5 max-sm:mx-auto max-sm:h-10 max-sm:w-10 max-sm:justify-center max-sm:px-0",
                )}
              >
                {isActive && (
                  <span
                    className={clsx(
                      "absolute bottom-2 top-2 w-[3px] rounded-r bg-[var(--color-accent)]",
                      sidebarCollapsed ? "-left-2" : "-left-3 max-sm:-left-2",
                    )}
                  />
                )}
                <Icon className="h-[18px] w-[18px] shrink-0 opacity-90" />
                {!sidebarCollapsed && (
                  <>
                    <span className="min-w-0 flex-1 truncate text-[13.5px] font-medium max-sm:hidden">
                      {item.name}
                    </span>
                    <span
                      className={clsx(
                        "font-mono text-[9.5px] max-sm:hidden",
                        isActive ? "text-[#7c8dff]" : "text-[#5a5e73]",
                      )}
                    >
                      {label}
                    </span>
                  </>
                )}
              </Link>
            );
          })}
        </div>
      </nav>

      <div className="py-3">
        <button
          type="button"
          onClick={toggleSidebar}
          className="mx-auto flex h-9 w-10 items-center justify-center rounded-lg text-[#8b8fa3] transition-colors hover:bg-white/[0.06] hover:text-[#e7e8f2]"
          title={sidebarCollapsed ? "展开侧边栏" : "收起侧边栏"}
        >
          {sidebarCollapsed ? (
            <ChevronRight className="h-[18px] w-[18px]" />
          ) : (
            <ChevronLeft className="h-[18px] w-[18px]" />
          )}
        </button>
      </div>
    </aside>
  );
}
