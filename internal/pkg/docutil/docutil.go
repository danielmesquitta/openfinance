package docutil

import (
	"errors"
	"fmt"
	"regexp"
	"strings"
)

const (
	cpfLength        = 11
	cnpjLength       = 14
	cpfFormattedLen  = 14 // XXX.XXX.XXX-XX
	cnpjFormattedLen = 18 // XX.XXX.XXX/XXXX-XX

	errMsgCNPJSeparator = "failed to write CNPJ separator: %w"
	errMsgCPFSeparator  = "failed to write CPF separator: %w"
)

func CleanDocument(doc string) string {
	re := regexp.MustCompile("[^0-9]")
	onlyDigitsDoc := re.ReplaceAllString(doc, "")

	return onlyDigitsDoc
}

func IsCPF(doc string) bool {
	cleanDoc := CleanDocument(doc)

	return len(cleanDoc) == cpfLength
}

func IsCNPJ(doc string) bool {
	cleanDoc := CleanDocument(doc)

	return len(cleanDoc) == cnpjLength
}

func MaskDocument(doc string) (string, error) {
	doc = CleanDocument(doc)

	switch len(doc) {
	case cnpjLength:
		return formatCNPJ(doc)
	case cpfLength:
		return formatCPF(doc)
	default:
		return "", errors.New("invalid document")
	}
}

func formatCNPJ(doc string) (string, error) {
	// CNPJ: XX.XXX.XXX/XXXX-XX
	var b strings.Builder

	b.Grow(cnpjFormattedLen)
	if _, err := b.WriteString(doc[:2]); err != nil {
		return "", fmt.Errorf("failed to write CNPJ part 1: %w", err)
	}
	if err := b.WriteByte('.'); err != nil {
		return "", fmt.Errorf(errMsgCNPJSeparator, err)
	}
	if _, err := b.WriteString(doc[2:5]); err != nil {
		return "", fmt.Errorf("failed to write CNPJ part 2: %w", err)
	}
	if err := b.WriteByte('.'); err != nil {
		return "", fmt.Errorf(errMsgCNPJSeparator, err)
	}
	if _, err := b.WriteString(doc[5:8]); err != nil {
		return "", fmt.Errorf("failed to write CNPJ part 3: %w", err)
	}
	if err := b.WriteByte('/'); err != nil {
		return "", fmt.Errorf(errMsgCNPJSeparator, err)
	}
	if _, err := b.WriteString(doc[8:12]); err != nil {
		return "", fmt.Errorf("failed to write CNPJ part 4: %w", err)
	}
	if err := b.WriteByte('-'); err != nil {
		return "", fmt.Errorf(errMsgCNPJSeparator, err)
	}
	if _, err := b.WriteString(doc[12:]); err != nil {
		return "", fmt.Errorf("failed to write CNPJ part 5: %w", err)
	}

	return b.String(), nil
}

func formatCPF(doc string) (string, error) {
	// CPF: XXX.XXX.XXX-XX
	var b strings.Builder

	b.Grow(cpfFormattedLen)
	if _, err := b.WriteString(doc[:3]); err != nil {
		return "", fmt.Errorf("failed to write CPF part 1: %w", err)
	}
	if err := b.WriteByte('.'); err != nil {
		return "", fmt.Errorf(errMsgCPFSeparator, err)
	}
	if _, err := b.WriteString(doc[3:6]); err != nil {
		return "", fmt.Errorf("failed to write CPF part 2: %w", err)
	}
	if err := b.WriteByte('.'); err != nil {
		return "", fmt.Errorf(errMsgCPFSeparator, err)
	}
	if _, err := b.WriteString(doc[6:9]); err != nil {
		return "", fmt.Errorf("failed to write CPF part 3: %w", err)
	}
	if err := b.WriteByte('-'); err != nil {
		return "", fmt.Errorf(errMsgCPFSeparator, err)
	}
	if _, err := b.WriteString(doc[9:]); err != nil {
		return "", fmt.Errorf("failed to write CPF part 4: %w", err)
	}

	return b.String(), nil
}
