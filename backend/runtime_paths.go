package backend

import "facade/backend/internal/apppath"

// EnsureRuntimeLayout 为运行时准备已安装应用的用户可写目录。
func EnsureRuntimeLayout(appRoot string) error {
	return apppath.EnsureWritableLayout(appRoot)
}

// ResolveRuntimePath 将相对路径解析到安装目录或用户状态目录。
func ResolveRuntimePath(appRoot, p string) string {
	return apppath.Resolve(appRoot, p)
}

// RuntimeStateRoot 返回当前运行时使用的状态目录。
func RuntimeStateRoot(appRoot string) string {
	return apppath.StateRoot(appRoot)
}

// RuntimeUsesDetachedState 表示是否使用独立于安装/项目目录的用户状态根目录。
func RuntimeUsesDetachedState(appRoot string) bool {
	return apppath.IsDetached(appRoot)
}
