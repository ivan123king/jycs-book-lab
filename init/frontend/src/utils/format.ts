// format.ts 提供页面展示层可复用的本地格式化函数。

/**
 * 将字节数格式化为书库页面可读的文件大小。
 */
export function formatBytes(bytes: number) {
  if (bytes <= 0) {
    return '0 B'
  }
  const units = ['B', 'KB', 'MB', 'GB']
  let size = bytes
  let unitIndex = 0
  while (size >= 1024 && unitIndex < units.length - 1) {
    size /= 1024
    unitIndex += 1
  }
  return `${size.toFixed(size >= 10 || unitIndex === 0 ? 0 : 1)} ${units[unitIndex]}`
}

/**
 * 将后端保存的 RFC3339 时间转换为中文本地展示文本。
 */
export function formatDateTime(value: string) {
  if (!value) {
    return '未知时间'
  }
  const date = new Date(value)
  if (Number.isNaN(date.getTime())) {
    return value
  }
  return date.toLocaleString('zh-CN', { hour12: false })
}

/**
 * 统一把 Wails 调用异常转换为页面可展示的错误文本。
 */
export function getErrorMessage(error: unknown) {
  if (error instanceof Error) {
    return error.message
  }
  return String(error)
}
