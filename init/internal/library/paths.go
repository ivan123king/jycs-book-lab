// paths.go 负责项目内路径归一化、安全校验和路径比较。
package library

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// samePath 用绝对路径做大小写不敏感比较，兼容 Windows 路径表现。
func samePath(left string, right string) bool {
	leftAbs, leftErr := filepath.Abs(filepath.Clean(left))
	rightAbs, rightErr := filepath.Abs(filepath.Clean(right))
	if leftErr == nil && rightErr == nil {
		return strings.EqualFold(leftAbs, rightAbs)
	}
	return strings.EqualFold(filepath.Clean(left), filepath.Clean(right))
}

// isPathInside 校验待操作路径位于指定父目录内部。
func isPathInside(parent string, child string) bool {
	parentAbs, err := filepath.Abs(filepath.Clean(parent))
	if err != nil {
		return false
	}
	childAbs, err := filepath.Abs(filepath.Clean(child))
	if err != nil {
		return false
	}
	relativePath, err := filepath.Rel(parentAbs, childAbs)
	if err != nil {
		return false
	}
	return relativePath != "." &&
		relativePath != ".." &&
		!strings.HasPrefix(relativePath, ".."+string(os.PathSeparator)) &&
		!filepath.IsAbs(relativePath)
}

// safeProjectPath 将项目内相对路径解析为安全的绝对路径，拒绝越界访问。
func safeProjectPath(bookPath string, relativePath string) (string, error) {
	root := projectRootPath(bookPath)
	cleanRelative := filepath.Clean(filepath.FromSlash(strings.TrimSpace(relativePath)))
	if cleanRelative == "." || filepath.IsAbs(cleanRelative) || strings.HasPrefix(cleanRelative, "..") {
		return "", fmt.Errorf("非法项目内路径: %s", relativePath)
	}
	fullPath := filepath.Join(root, cleanRelative)
	rel, err := filepath.Rel(root, fullPath)
	if err != nil || strings.HasPrefix(rel, "..") || filepath.IsAbs(rel) {
		return "", fmt.Errorf("路径越界: %s", relativePath)
	}
	return fullPath, nil
}

// projectRootPath 归一化单本小说项目根目录路径。
func projectRootPath(bookPath string) string {
	return filepath.Clean(strings.TrimSpace(bookPath))
}
