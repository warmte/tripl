package document

import (
	"context"
	"errors"
	"fmt"
	"io"
	"strings"

	pdf "github.com/ledongthuc/pdf"
	. "github.com/warmte/tripl/common"
)

const (
	MaxPDFTextLength = 3000
)

type PDFParser struct{}

func NewPDFParser() *PDFParser {
	return &PDFParser{}
}

var ErrPDFTextLimitReached = errors.New("pdf text limit reached")

func (p *PDFParser) Parse(
	ctx context.Context,
	document *PDFDocument,
) (string, error) {
	if document == nil {
		return "", errors.New("pdf document is nil")
	}

	if document.Path == "" {
		return "", errors.New("pdf document path is empty")
	}

	if err := ctx.Err(); err != nil {
		return "", err
	}

	file, reader, err := pdf.Open(document.Path)
	if err != nil {
		return "", fmt.Errorf("open pdf: %w", err)
	}
	defer file.Close()

	textReader, err := reader.GetPlainText()
	if err != nil {
		return "", fmt.Errorf("extract pdf text: %w", err)
	}

	limitedReader := io.LimitReader(textReader, MaxPDFTextLength+1)

	data, err := io.ReadAll(limitedReader)
	if err != nil {
		return "", fmt.Errorf("read extracted pdf text: %w", err)
	}

	if len(data) > MaxPDFTextLength {
		data = data[:MaxPDFTextLength]
	}

	result := strings.TrimSpace(string(data))

	if result == "" {
		return "", ErrEmptyFile
	}

	return result, nil
}
