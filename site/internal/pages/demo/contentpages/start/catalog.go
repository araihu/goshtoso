package startpages

import "github.com/araihu/goshtoso/site/internal/pages/catalog"

func componentCount() int {
	return len(catalog.ComponentPages())
}
