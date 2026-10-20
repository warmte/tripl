package document

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os/exec"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	. "github.com/warmte/tripl/common"
)

const (
	MaxPDFTextLength = 3000
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

	parseCtx, cancel := context.WithTimeout(ctx, pdfParseTimeout)
	defer cancel()

	cmd := exec.CommandContext(
		parseCtx,
		"pdftotext",
		"-l", strconv.Itoa(MaxPDFPages),
		"-enc", "UTF-8",
		document.Path,
		"-",
	)

	var output bytes.Buffer
	cmd.Stdout = &output

	if err := cmd.Run(); err != nil {
		if parseCtx.Err() != nil {
			return "", fmt.Errorf(
				"pdftotext timed out or was canceled: %w",
				parseCtx.Err(),
			)
		}

		return "", fmt.Errorf("run pdftotext: %w", err)
	}

	text := strings.TrimSpace(output.String())

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
