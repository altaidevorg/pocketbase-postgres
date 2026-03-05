package search

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/pocketbase/dbx"
	"github.com/pocketbase/pocketbase/tools/inflector"
	"github.com/pocketbase/pocketbase/tools/list"
)

type NullFallbackPreference int

const (
	NullFallbackAuto     NullFallbackPreference = 0
	NullFallbackDisabled NullFallbackPreference = 1
	NullFallbackEnforced NullFallbackPreference = 2
)

// ResolverResult defines a single FieldResolver.Resolve() successfully parsed result.
type ResolverResult struct {
	// Identifier is the plain SQL identifier/column that will be used
	// in the final db expression as left or right operand.
	Identifier string

	// NullFallback specify the preference for how NULL or empty values
	// should be resolved (default to "auto").
	//
	// Set to NullFallbackDisabled to prevent any COALESCE or NULL fallbacks.
	// Set to NullFallbackEnforced to prefer COALESCE or NULL fallbacks when needed.
	NullFallback NullFallbackPreference

	// Params is a map with db placeholder->value pairs that will be added
	// to the query when building both resolved operands/sides in a single expression.
	Params dbx.Params

	// MultiMatchSubQuery is an optional sub query expression that will be added
	// in addition to the combined ResolverResult expression during build.
	MultiMatchSubQuery *MultiMatchSubquery

	// AfterBuild is an optional function that will be called after building
	// and combining the result of both resolved operands/sides in a single expression.
	AfterBuild func(expr dbx.Expression) dbx.Expression
}

// FieldResolver defines an interface for managing search fields.
type FieldResolver interface {
	// UpdateQuery allows to updated the provided db query based on the
	// resolved search fields (eg. adding joins aliases, etc.).
	//
	// Called internally by `search.Provider` before executing the search request.
	UpdateQuery(query *dbx.SelectQuery) error

	// Resolve parses the provided field and returns a properly
	// formatted db identifier (eg. NULL, quoted column, placeholder parameter, etc.).
	Resolve(field string) (*ResolverResult, error)
}

// NewSimpleFieldResolver creates a new `SimpleFieldResolver` with the
// provided `allowedFields`.
//
// Each `allowedFields` could be a plain string (eg. "name")
// or a regexp pattern (eg. `^\w+[\w\.]*$`).
func NewSimpleFieldResolver(allowedFields ...string) *SimpleFieldResolver {
	return &SimpleFieldResolver{
		allowedFields: allowedFields,
	}
}

// SimpleFieldResolver defines a generic search resolver that allows
// only its listed fields to be resolved and take part in a search query.
//
// If `allowedFields` are empty no fields filtering is applied.
type SimpleFieldResolver struct {
	allowedFields []string
	isPostgres    bool
}

// SetIsPostgres configures the resolver to generate PostgreSQL compatible
// queries (eg. using "->" instead of "JSON_EXTRACT").
func (r *SimpleFieldResolver) SetIsPostgres(v bool) *SimpleFieldResolver {
	r.isPostgres = v
	return r
}

// IsPostgres settings of the resolver.
func (r *SimpleFieldResolver) IsPostgres() bool {
	return r.isPostgres
}

// UpdateQuery implements `search.UpdateQuery` interface.
func (r *SimpleFieldResolver) UpdateQuery(query *dbx.SelectQuery) error {
	// nothing to update...
	return nil
}

// Resolve implements `search.Resolve` interface.
//
// Returns error if `field` is not in `r.allowedFields`.
func (r *SimpleFieldResolver) Resolve(field string) (*ResolverResult, error) {
	if !list.ExistInSliceWithRegex(field, r.allowedFields) {
		return nil, fmt.Errorf("failed to resolve field %q", field)
	}

	parts := strings.Split(field, ".")

	// single regular field
	if len(parts) == 1 {
		return &ResolverResult{
			Identifier: "[[" + inflector.Columnify(parts[0]) + "]]",
		}, nil
	}

	// treat as json path
	var jsonPath strings.Builder
	if r.isPostgres {
		// Postgres uses -> for json path traversal and ->> for checks
		// -----------------------------------------------------------
		jsonPath.WriteString(inflector.Columnify(parts[0]))
		for i, part := range parts[1:] {
			if _, err := strconv.Atoi(part); err == nil {
				jsonPath.WriteString("->")
				jsonPath.WriteString(part)
			} else {
				jsonPath.WriteString("->'")
				jsonPath.WriteString(inflector.Columnify(part))
				jsonPath.WriteString("'")
			}
			// if it is the last part, use ->> to unquote the result
			if i == len(parts[1:])-1 {
				jsonPath.WriteString(">>") // fix: the loop constructs ->'key' or ->index, so we need to change the last arrow to ->>
			}
		}

		// Rewrite the construction to be cleaner for Postgres:
		// col->'a'->'b' ->> 'c'
		jsonPath.Reset()
		jsonPath.WriteString("[[")
		jsonPath.WriteString(inflector.Columnify(parts[0]))
		jsonPath.WriteString("]]")
		for i, part := range parts[1:] {
			isLast := i == len(parts[1:])-1
			if _, err := strconv.Atoi(part); err == nil {
				if isLast {
					jsonPath.WriteString("->>")
				} else {
					jsonPath.WriteString("->")
				}
				jsonPath.WriteString(part)
			} else {
				if isLast {
					jsonPath.WriteString("->>'")
				} else {
					jsonPath.WriteString("->'")
				}
				jsonPath.WriteString(inflector.Columnify(part))
				jsonPath.WriteString("'")
			}
		}

		return &ResolverResult{
			NullFallback: NullFallbackDisabled,
			Identifier:   jsonPath.String(),
		}, nil
	}

	// SQLite uses JSON_EXTRACT
	// -----------------------------------------------------------
	jsonPath.WriteString("$")
	for _, part := range parts[1:] {
		if _, err := strconv.Atoi(part); err == nil {
			jsonPath.WriteString("[")
			jsonPath.WriteString(inflector.Columnify(part))
			jsonPath.WriteString("]")
		} else {
			jsonPath.WriteString(".")
			jsonPath.WriteString(inflector.Columnify(part))
		}
	}

	return &ResolverResult{
		NullFallback: NullFallbackDisabled,
		Identifier: fmt.Sprintf(
			"JSON_EXTRACT([[%s]], '%s')",
			inflector.Columnify(parts[0]),
			jsonPath.String(),
		),
	}, nil
}
