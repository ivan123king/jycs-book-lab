// storage.go 负责 library 包内 JSON、TXT 和导入源文件的本地持久化工具。
package library

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"unicode/utf8"

	"golang.org/x/text/encoding/simplifiedchinese"
)

// readOptionalJSONFile 读取可选 JSON 文件，调用方决定缺失文件是否属于错误。
func readOptionalJSONFile(path string, target any) error {
	content, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	if err := json.Unmarshal(content, target); err != nil {
		return err
	}
	return nil
}

// readNovelText 读取小说 TXT，并兼容 UTF-8 和 GB18030 编码。
func readNovelText(path string) (string, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("读取小说文本失败: %w", err)
	}
	content = trimUTF8BOM(content)
	if utf8.Valid(content) {
		return string(content), nil
	}
	decoded, err := simplifiedchinese.GB18030.NewDecoder().Bytes(content)
	if err != nil {
		return "", fmt.Errorf("小说文本编码识别失败，请先转换为 UTF-8 或 GB18030 后重试: %w", err)
	}
	return string(decoded), nil
}

// trimUTF8BOM 去掉 UTF-8 BOM，避免正文或 JSON 字段前出现不可见字符。
func trimUTF8BOM(content []byte) []byte {
	if len(content) >= 3 && content[0] == 0xEF && content[1] == 0xBB && content[2] == 0xBF {
		return content[3:]
	}
	return content
}

// copyFile 将导入源文件复制到小说项目目录，保留原始导入文本。
func copyFile(source string, target string) error {
	input, err := os.Open(source)
	if err != nil {
		return fmt.Errorf("打开导入文件失败: %w", err)
	}
	defer input.Close()

	output, err := os.Create(target)
	if err != nil {
		return fmt.Errorf("创建原文副本失败: %w", err)
	}
	defer output.Close()

	if _, err := io.Copy(output, input); err != nil {
		return fmt.Errorf("复制原文副本失败: %w", err)
	}
	return output.Sync()
}

// writeJSONFileAtomic 通过临时文件替换写入 JSON，降低半写入导致数据损坏的风险。
func writeJSONFileAtomic(path string, value any) error {
	content, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		return fmt.Errorf("序列化 JSON 失败: %w", err)
	}
	content = append(content, '\n')

	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return fmt.Errorf("创建 JSON 目录失败: %w", err)
	}
	tempFile, err := os.CreateTemp(filepath.Dir(path), filepath.Base(path)+".*.tmp")
	if err != nil {
		return fmt.Errorf("创建临时 JSON 文件失败: %w", err)
	}
	tempPath := tempFile.Name()
	if _, err := tempFile.Write(content); err != nil {
		tempFile.Close()
		_ = os.Remove(tempPath)
		return fmt.Errorf("写入临时 JSON 文件失败: %w", err)
	}
	if err := tempFile.Close(); err != nil {
		_ = os.Remove(tempPath)
		return fmt.Errorf("关闭临时 JSON 文件失败: %w", err)
	}
	if err := os.Rename(tempPath, path); err != nil {
		if _, statErr := os.Stat(path); statErr != nil {
			_ = os.Remove(tempPath)
			return fmt.Errorf("替换 JSON 文件失败: %w", err)
		}
		if removeErr := os.Remove(path); removeErr != nil {
			_ = os.Remove(tempPath)
			return fmt.Errorf("删除旧 JSON 文件失败: %w", removeErr)
		}
		if renameErr := os.Rename(tempPath, path); renameErr != nil {
			_ = os.Remove(tempPath)
			return fmt.Errorf("替换 JSON 文件失败: %w", renameErr)
		}
	}
	return nil
}

// readJSONFile 读取并解析必需 JSON 文件。
func readJSONFile(path string, target any) error {
	content, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	if err := json.Unmarshal(content, target); err != nil {
		return err
	}
	return nil
}

// writeTextFileAtomic 通过临时文件替换写入 TXT 正文内容。
func writeTextFileAtomic(path string, value string) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return fmt.Errorf("创建文本目录失败: %w", err)
	}
	tempFile, err := os.CreateTemp(filepath.Dir(path), filepath.Base(path)+".*.tmp")
	if err != nil {
		return fmt.Errorf("创建临时文本文件失败: %w", err)
	}
	tempPath := tempFile.Name()
	if _, err := tempFile.WriteString(value); err != nil {
		tempFile.Close()
		_ = os.Remove(tempPath)
		return fmt.Errorf("写入临时文本文件失败: %w", err)
	}
	if err := tempFile.Close(); err != nil {
		_ = os.Remove(tempPath)
		return fmt.Errorf("关闭临时文本文件失败: %w", err)
	}
	if err := os.Rename(tempPath, path); err != nil {
		if removeErr := os.Remove(path); removeErr != nil && !errors.Is(removeErr, os.ErrNotExist) {
			_ = os.Remove(tempPath)
			return fmt.Errorf("删除旧文本文件失败: %w", removeErr)
		}
		if renameErr := os.Rename(tempPath, path); renameErr != nil {
			_ = os.Remove(tempPath)
			return fmt.Errorf("替换文本文件失败: %w", renameErr)
		}
	}
	return nil
}
