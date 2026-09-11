package docspages

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strconv"

	"github.com/a-h/templ"
	"github.com/araihu/goshtoso/expressions"
)

type expressionDemoKey struct{}
type expressionDemoState struct {
	Language string
	Page     int
}

// WithExpressionDemo reads this demo's request-owned language and page selection.
func WithExpressionDemo(r *http.Request) *http.Request {
	language := r.URL.Query().Get("language")
	if language != "pt" && language != "es" && language != "zh" && language != "de" {
		language = "en"
	}
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	if page < 1 || page > 5 {
		page = 2
	}
	return r.WithContext(context.WithValue(r.Context(), expressionDemoKey{}, expressionDemoState{language, page}))
}

func expressionDemoFrom(ctx context.Context) expressionDemoState {
	if state, ok := ctx.Value(expressionDemoKey{}).(expressionDemoState); ok {
		return state
	}
	return expressionDemoState{"en", 2}
}

func translatedDemoSet(language string) expressions.Set {
	switch language {
	case "pt":
		return expressions.Set{
			EmptyState: expressions.EmptyState{Title: "Ainda não há nada aqui", Description: "Os itens aparecerão aqui quando estiverem disponíveis."},
			FileInput:  expressions.FileInput{BrowseLabel: "Procurar", DropText: " ou arraste e solte aqui", EmptyText: "Nenhum arquivo selecionado"},
			CodeBlock:  expressions.CodeBlock{CopyLabel: "Copiar", CopiedLabel: "Copiado!", ErrorText: "Não foi possível copiar", CopyAriaLabel: func(label string) string { return "Copiar " + label }},
			Pagination: expressions.Pagination{AriaLabel: "Paginação", PreviousLabel: "Anterior", NextLabel: "Próxima", PreviousAriaLabel: "Página anterior", NextAriaLabel: "Próxima página", MoreAriaLabel: "Mais páginas", PageAriaLabel: func(page int) string { return fmt.Sprintf("Página %d", page) }},
		}
	case "es":
		return expressions.Set{
			EmptyState: expressions.EmptyState{Title: "Todavía no hay nada aquí", Description: "Los elementos aparecerán aquí cuando estén disponibles."},
			FileInput:  expressions.FileInput{BrowseLabel: "Examinar", DropText: " o arrastra y suelta aquí", EmptyText: "Ningún archivo seleccionado"},
			CodeBlock:  expressions.CodeBlock{CopyLabel: "Copiar", CopiedLabel: "¡Copiado!", ErrorText: "No se pudo copiar", CopyAriaLabel: func(label string) string { return "Copiar " + label }},
			Pagination: expressions.Pagination{AriaLabel: "Paginación", PreviousLabel: "Anterior", NextLabel: "Siguiente", PreviousAriaLabel: "Página anterior", NextAriaLabel: "Página siguiente", MoreAriaLabel: "Más páginas", PageAriaLabel: func(page int) string { return fmt.Sprintf("Página %d", page) }},
		}
	case "zh":
		return expressions.Set{
			EmptyState: expressions.EmptyState{Title: "这里还没有内容", Description: "有可用项目时，它们会显示在这里。"},
			FileInput:  expressions.FileInput{BrowseLabel: "浏览文件", DropText: " 或将文件拖放到这里", EmptyText: "未选择文件"},
			CodeBlock:  expressions.CodeBlock{CopyLabel: "复制", CopiedLabel: "已复制！", ErrorText: "无法复制", CopyAriaLabel: func(label string) string { return "复制" + label }},
			Pagination: expressions.Pagination{AriaLabel: "分页", PreviousLabel: "上一页", NextLabel: "下一页", PreviousAriaLabel: "上一页", NextAriaLabel: "下一页", MoreAriaLabel: "更多页", PageAriaLabel: func(page int) string { return fmt.Sprintf("第 %d 页", page) }},
		}
	case "de":
		return expressions.Set{
			EmptyState: expressions.EmptyState{Title: "Hier gibt es noch nichts", Description: "Elemente werden hier angezeigt, sobald sie verfügbar sind."},
			FileInput:  expressions.FileInput{BrowseLabel: "Durchsuchen", DropText: " oder Dateien hierher ziehen", EmptyText: "Keine Datei ausgewählt"},
			CodeBlock:  expressions.CodeBlock{CopyLabel: "Kopieren", CopiedLabel: "Kopiert!", ErrorText: "Kopieren fehlgeschlagen", CopyAriaLabel: func(label string) string { return label + " kopieren" }},
			Pagination: expressions.Pagination{AriaLabel: "Seitennavigation", PreviousLabel: "Zurück", NextLabel: "Weiter", PreviousAriaLabel: "Vorherige Seite", NextAriaLabel: "Nächste Seite", MoreAriaLabel: "Weitere Seiten", PageAriaLabel: func(page int) string { return fmt.Sprintf("Seite %d", page) }},
		}
	default:
		return expressions.Set{}
	}
}

func expressionPreview(language string, request bool, child templ.Component) templ.Component {
	return templ.ComponentFunc(func(ctx context.Context, w io.Writer) error {
		set := translatedDemoSet(language)
		if request {
			return child.Render(expressions.With(ctx, set), w)
		}
		return expressions.NewRenderer(set).Render(ctx, w, child)
	})
}
