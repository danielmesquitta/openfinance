package docutil

import (
	"errors"
	"regexp"
	"strings"
)

const (
	cpfLength        = 11
	cnpjLength       = 14
	cpfFormattedLen  = 14 // XXX.XXX.XXX-XX
	cnpjFormattedLen = 18 // XX.XXX.XXX/XXXX-XX
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
		// CNPJ: XX.XXX.XXX/XXXX-XX
		var b strings.Builder
		b.Grow(cnpjFormattedLen)
		b.WriteString(doc[:2])
		b.WriteByte('.')
		b.WriteString(doc[2:5])
		b.WriteByte('.')
		b.WriteString(doc[5:8])
		b.WriteByte('/')
		b.WriteString(doc[8:12])
		b.WriteByte('-')
		b.WriteString(doc[12:])
		return b.String(), nil

	case cpfLength:
		// CPF: XXX.XXX.XXX-XX
		var b strings.Builder
		b.Grow(cpfFormattedLen)
		b.WriteString(doc[:3])
		b.WriteByte('.')
		b.WriteString(doc[3:6])
		b.WriteByte('.')
		b.WriteString(doc[6:9])
		b.WriteByte('-')
		b.WriteString(doc[9:])
		return b.String(), nil

	default:
		return "", errors.New("invalid document")
	}
}
