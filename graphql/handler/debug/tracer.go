package debug

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"strings"

	"github.com/mattn/go-colorable"

	"github.com/99designs/gqlgen/graphql"
)

type Tracer struct {
	DisableColor bool
	out          io.Writer
}

var _ interface {
	graphql.HandlerExtension
	graphql.ResponseInterceptor
} = &Tracer{}

func (a Tracer) ExtensionName() string {
	return "ApolloTracing"
}

func (a *Tracer) Validate(schema graphql.ExecutableSchema) error {
	a.out = colorable.NewColorableStdout()

	return nil
}

// colorize wraps s in the ANSI SGR sequence code, the same bytes that aurora/v4 writes.
func colorize(code, s string) string {
	return "\033[" + code + "m" + s + "\033[0m"
}

func stringify(value any) string {
	valueJson, err := json.MarshalIndent(value, "  ", "  ")
	if err == nil {
		return string(valueJson)
	}

	return fmt.Sprint(value)
}

func (a Tracer) InterceptResponse(ctx context.Context, next graphql.ResponseHandler) *graphql.Response {
	opCtx := graphql.GetOperationContext(ctx)

	_, _ = fmt.Fprintln(a.out, "GraphQL Request {")
	for _, line := range strings.Split(opCtx.RawQuery, "\n") {
		_, _ = fmt.Fprintln(a.out, " ", colorize("36", line))
	}
	for name, value := range opCtx.Variables {
		_, _ = fmt.Fprintf(a.out, "  var %s = %s\n", name, colorize("33", stringify(value)))
	}
	resp := next(ctx)

	_, _ = fmt.Fprintln(a.out, "  resp:", colorize("32", stringify(resp)))
	if resp != nil {
		for _, err := range resp.Errors {
			_, _ = fmt.Fprintln(a.out, "  error:", colorize("1", err.Path.String()+":"), colorize("31", err.Message))
		}
	}

	_, _ = fmt.Fprintln(a.out, "}")
	_, _ = fmt.Fprintln(a.out)
	return resp
}
