package docutil

import (
	"regexp"

	"github.com/danielmesquitta/openfinance/internal/domain/errs"
)

func CleanDocument(doc string) string {
	re := regexp.MustCompile("[^0-9]")
	onlyDigitsDoc := re.ReplaceAllString(doc, "")
	return onlyDigitsDoc
}

func IsCPF(doc string) bool {
	cleanDoc := CleanDocument(doc)
	return len(cleanDoc) == 11
}

func IsCNPJ(doc string) bool {
	cleanDoc := CleanDocument(doc)
	return len(cleanDoc) == 14
}

func MaskDocument(doc string) (string, error) {
	doc = CleanDocument(doc)

	if IsCNPJ(doc) {
		return doc[:2] + "." + doc[2:5] + "." + doc[5:8] + "/" + doc[8:12] + "-" + doc[12:], nil
	} else if IsCPF(doc) {
		return doc[:3] + "." + doc[3:6] + "." + doc[6:9] + "-" + doc[9:], nil
	}

	return "", errs.New("invalid document")
}
