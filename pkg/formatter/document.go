package formatter

import (
	"errors"
	"strings"
)

func MaskDocument(doc string, docType string) (string, error) {
	cleanDoc := strings.ReplaceAll(doc, ".", "")
	cleanDoc = strings.ReplaceAll(cleanDoc, "-", "")
	cleanDoc = strings.ReplaceAll(cleanDoc, "/", "")

	switch docType {
	case "CPF":
		if len(cleanDoc) != 11 {
			return "", errors.New("Invalid CPF")
		}
		return cleanDoc[:3] + "." + cleanDoc[3:6] + "." + cleanDoc[6:9] + "-" + cleanDoc[9:], nil
	case "CNPJ":
		if len(cleanDoc) != 14 {
			return "", errors.New("Invalid CNPJ")
		}
		return cleanDoc[:2] + "." + cleanDoc[2:5] + "." + cleanDoc[5:8] + "/" + cleanDoc[8:12] + "-" + cleanDoc[12:], nil
	default:
		return "", errors.New("Unknown Document Type")
	}
}
