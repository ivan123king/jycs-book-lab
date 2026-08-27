// settings_service.go 负责应用级设置读写和最近打开小说入口维护。
package library

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// loadSettings 从应用系统数据目录读取全局设置，不读取或写入单本小说业务数据。
func loadSettings() (AppSettings, error) {
	settings := AppSettings{Version: 1}
	path, err := settingsPath()
	if err != nil {
		return settings, err
	}

	content, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		return normalizeAppSettings(settings), nil
	}
	if err != nil {
		return settings, fmt.Errorf("读取应用设置失败: %w", err)
	}
	if err := json.Unmarshal(content, &settings); err != nil {
		return settings, fmt.Errorf("解析应用设置失败: %w", err)
	}
	return normalizeAppSettings(settings), nil
}

// saveSettings 将书库目录等全局偏好保存到应用系统数据目录。
func saveSettings(settings AppSettings) error {
	settings = normalizeAppSettings(settings)
	path, err := settingsPath()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return fmt.Errorf("创建应用设置目录失败: %w", err)
	}
	return writeJSONFileAtomic(path, settings)
}

// LoadAppSettings 返回设置页可编辑的应用级偏好，不包含任何单本书正式业务数据。
func LoadAppSettings() (AppSettings, error) {
	return loadSettings()
}

// SaveAppSettings 保存应用级默认阅读字号；已有小说仍读取各自项目目录中的字号偏好。
func SaveAppSettings(request SaveAppSettingsRequest) (AppSettings, error) {
	settings, err := loadSettings()
	if err != nil {
		return AppSettings{}, err
	}
	settings.DefaultReaderFontSize = request.DefaultReaderFontSize
	if err := saveSettings(settings); err != nil {
		return AppSettings{}, err
	}
	return loadSettings()
}

// defaultReadingPreferencesForNewBook 读取新导入小说的初始阅读偏好，之后随单书项目目录迁移。
func defaultReadingPreferencesForNewBook() (ReadingPreferences, error) {
	settings, err := loadSettings()
	if err != nil {
		return ReadingPreferences{}, err
	}
	return ReadingPreferences{FontSize: settings.DefaultReaderFontSize}, nil
}

// normalizeAppSettings 补齐旧设置文件缺失的默认值，并限制字号在阅读页支持范围内。
func normalizeAppSettings(settings AppSettings) AppSettings {
	if settings.Version == 0 {
		settings.Version = 1
	}
	if settings.DefaultReaderFontSize == 0 {
		settings.DefaultReaderFontSize = defaultReaderFontSize
	}
	settings.DefaultReaderFontSize = clampInt(settings.DefaultReaderFontSize, minReaderFontSize, maxReaderFontSize)
	return settings
}

// settingsPath 返回应用级设置文件路径，只用于全局偏好。
func settingsPath() (string, error) {
	configDir, err := os.UserConfigDir()
	if err != nil {
		return "", fmt.Errorf("获取应用设置目录失败: %w", err)
	}
	return filepath.Join(configDir, appDataFolderName, settingsFileName), nil
}

// RememberLastOpenedBookPath 只把最近打开的小说目录入口写入应用设置，阅读位置仍以单书目录为准。
func RememberLastOpenedBookPath(bookPath string) error {
	bookPath = filepath.Clean(strings.TrimSpace(bookPath))
	if bookPath == "." || bookPath == "" {
		return errors.New("最近打开的小说目录不能为空")
	}
	if _, err := OpenBookProject(bookPath); err != nil {
		return err
	}
	settings, err := loadSettings()
	if err != nil {
		return err
	}
	settings.LastOpenedBookPath = bookPath
	return saveSettings(settings)
}

// OpenLastReadingSession 根据应用设置中的最近小说入口恢复阅读工作台。
func OpenLastReadingSession() (*ReadingSession, error) {
	settings, err := loadSettings()
	if err != nil {
		return nil, err
	}
	if strings.TrimSpace(settings.LastOpenedBookPath) == "" {
		return nil, nil
	}
	session, err := OpenReadingSession(settings.LastOpenedBookPath)
	if err != nil {
		return nil, err
	}
	return &session, nil
}

// clearLastOpenedBookPathIfMatches 在书籍被移除后清理对应的最近打开入口。
func clearLastOpenedBookPathIfMatches(bookPath string) error {
	settings, err := loadSettings()
	if err != nil {
		return err
	}
	if !samePath(settings.LastOpenedBookPath, bookPath) {
		return nil
	}
	settings.LastOpenedBookPath = ""
	return saveSettings(settings)
}
