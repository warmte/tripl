package document

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	. "github.com/warmte/tripl/common"
)

type PDFService struct {
	api *tgbotapi.BotAPI
}

func NewPDFService(api *tgbotapi.BotAPI) *PDFService {
	return &PDFService{
		api: api,
	}
}

type PDFDocument struct {
	Path     string
	FileName string
	Size     int64
}

func (s *PDFService) Download(
	ctx context.Context,
	document *tgbotapi.Document,
) (*PDFDocument, error) {
	if document == nil {
		return nil, errors.New("document is nil")
	}

	if !isPDF(document) {
		return nil, ErrIncorrectFileFormat
	}

	if document.FileSize > MaxPDFSize {
		return nil, ErrTooLargeFile
	}

	if err := ctx.Err(); err != nil {
		return nil, err
	}

	file, err := s.api.GetFile(tgbotapi.FileConfig{FileID: document.FileID})
	if err != nil {
		return nil, fmt.Errorf("get telegram file: %w", err)
	}

	if err := ctx.Err(); err != nil {
		return nil, err
	}

	tmpDir, err := os.MkdirTemp("", "tripl-pdf-*")
	if err != nil {
		return nil, fmt.Errorf("create temp dir: %w", err)
	}

	path := filepath.Join(tmpDir, safeFileName(document.FileName))

	if err := s.download(ctx, file, path); err != nil {
		_ = os.RemoveAll(tmpDir)
		return nil, fmt.Errorf("download pdf: %w", err)
	}

	info, err := os.Stat(path)
	if err != nil {
		_ = os.RemoveAll(tmpDir)
		return nil, fmt.Errorf("stat downloaded pdf: %w", err)
	}

	if info.Size() > MaxPDFSize {
		_ = os.RemoveAll(tmpDir)
		return nil, ErrTooLargeFile
	}

	return &PDFDocument{
		Path:     path,
		FileName: document.FileName,
		Size:     info.Size(),
	}, nil
}

func (s *PDFService) Remove(document *PDFDocument) {
	if document == nil || document.Path == "" {
		return
	}

	_ = os.RemoveAll(filepath.Dir(document.Path))
}

func (s *PDFService) download(
	ctx context.Context,
	file tgbotapi.File,
	destination string,
) error {
	url := file.Link(s.api.Token)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return fmt.Errorf("create download request: %w", err)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("download request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf(
			"telegram returned status %s",
			resp.Status,
		)
	}

	output, err := os.OpenFile(destination, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0600)
	if err != nil {
		return fmt.Errorf("create destination file: %w", err)
	}
	defer output.Close()

	reader := io.LimitReader(resp.Body, MaxPDFSize+1)

	written, err := io.Copy(output, reader)
	if err != nil {
		return fmt.Errorf("save downloaded pdf: %w", err)
	}

	if written > MaxPDFSize {
		return ErrTooLargeFile
	}

	return nil
}

func isPDF(document *tgbotapi.Document) bool {
	if document == nil {
		return false
	}

	if strings.EqualFold(
		document.MimeType,
		"application/pdf",
	) {
		return true
	}

	return strings.EqualFold(
		filepath.Ext(document.FileName),
		".pdf",
	)
}

func safeFileName(name string) string {
	name = filepath.Base(name)

	if name == "" || name == "." || name == string(filepath.Separator) {
		return "document.pdf"
	}

	return name
}
