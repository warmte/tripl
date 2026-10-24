package document

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os/exec"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	. "github.com/warmte/tripl/common"
)

const (
	MaxPDFTextLength = 2800
	pdfParseTimeout  = 20 * time.Second
)

type PDFParser struct{}

func NewPDFParser() *PDFParser {
	return &PDFParser{}
}

func (p *PDFParser) Parse(ctx context.Context, document *PDFDocument) (string, error) {
	if document == nil {
		return "", errors.New("pdf document is nil")
	}

	if document.Path == "" {
		return "", errors.New("pdf document path is empty")
	}

	if err := ctx.Err(); err != nil {
		return "", err
	}

	output, err := runPDFParser(ctx, document.Path)
	if err != nil {
		return "", err
	}

	text := strings.TrimSpace(string(output))

	if !utf8.ValidString(text) {
		return "", errors.New("pdftotext returned invalid UTF-8")
	}

	runes := []rune(text)
	if len(runes) > MaxPDFTextLength {
		text = string(runes[:MaxPDFTextLength])
	}

	text = strings.TrimSpace(text)

	if text == "" {
		return "", ErrEmptyFile
	}

	return text, nil
}

func runPDFParser(ctx context.Context, path string) ([]byte, error) {
	parseCtx, cancel := context.WithTimeout(ctx, pdfParseTimeout)
	defer cancel()

	cmd := exec.CommandContext(
		parseCtx,
		"pdftotext",
		"-l", strconv.Itoa(MaxPDFPages),
		"-enc", "UTF-8",
		path,
		"-",
	)

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, fmt.Errorf("create parser stdout: %w", err)
	}

	if err := cmd.Start(); err != nil {
		return nil, fmt.Errorf("start PDF parser: %w", err)
	}

	output, readErr := io.ReadAll(io.LimitReader(stdout, MaxPDFSize+1))

	if int64(len(output)) > MaxPDFSize {
		_ = cmd.Process.Kill()
		_ = cmd.Wait()
		return nil, ErrTooLargeFile
	}

	if readErr != nil {
		_ = cmd.Process.Kill()
		_ = cmd.Wait()
		return nil, fmt.Errorf("read parser output: %w", readErr)
	}

	if err := cmd.Wait(); err != nil {
		return nil, fmt.Errorf("PDF parser failed: %w", err)
	}

	return output, nil
}
