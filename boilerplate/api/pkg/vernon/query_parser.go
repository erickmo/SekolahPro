package vernon

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
)

const (
	defaultLimit = 20
	minLimit     = 1
)

// ParseQueryParams mengurai query string HTTP request menjadi FindParams.
// Mendukung dot notation untuk nested JSONB fields: customer.name=Maju.
// Parameter sistem: _sort, _page, _limit.
func ParseQueryParams(r *http.Request, maxLimit int) FindParams {
	q := r.URL.Query()
	p := FindParams{
		Filters: make(map[string]string),
		Page:    1,
		Limit:   defaultLimit,
	}

	for key, vals := range q {
		if len(vals) == 0 {
			continue
		}
		switch key {
		case "_sort":
			p.Sort = vals[0]
		case "_page":
			if n, err := strconv.Atoi(vals[0]); err == nil && n > 0 {
				p.Page = n
			}
		case "_limit":
			if n, err := strconv.Atoi(vals[0]); err == nil && n >= minLimit {
				p.Limit = min(n, maxLimit)
			}
		default:
			p.Filters[key] = vals[0]
		}
	}

	return p
}

// BuildWhereClause membangun klausa WHERE + args dari scope dan filters.
// tenant_id dan company_id selalu di-inject sebagai filter pertama (keamanan).
// Filter biasa diterjemahkan ke JSONB path query.
func BuildWhereClause(tenantID, companyID string, filters map[string]string) (string, []any) {
	conditions := []string{
		"deleted_at IS NULL",
		"tenant_id = $1",
		"company_id = $2",
	}
	args := []any{tenantID, companyID}
	argIdx := 3

	for field, value := range filters {
		cond, arg := buildCondition(field, value, argIdx)
		conditions = append(conditions, cond)
		args = append(args, arg)
		argIdx++
	}

	return strings.Join(conditions, " AND "), args
}

// buildCondition membangun satu kondisi SQL dari field dan value.
// Mendukung dot notation: "customer.name" → `_data->'customer'->>'name'`
func buildCondition(field, value string, argIdx int) (string, any) {
	jsonPath := toJSONBPath(field)
	return fmt.Sprintf("%s ILIKE $%d", jsonPath, argIdx), "%" + value + "%"
}

// toJSONBPath mengubah field name menjadi PostgreSQL JSONB path expression.
// "status"         → `_data->>'status'`
// "customer.name"  → `_data->'customer'->>'name'`
func toJSONBPath(field string) string {
	parts := strings.Split(field, ".")
	if len(parts) == 1 {
		return fmt.Sprintf("_data->>'%s'", parts[0])
	}

	var sb strings.Builder
	sb.WriteString("_data")
	for i, p := range parts {
		if i == len(parts)-1 {
			fmt.Fprintf(&sb, "->>'%s'", p)
		} else {
			fmt.Fprintf(&sb, "->'%s'", p)
		}
	}
	return sb.String()
}

// BuildOrderClause membangun klausa ORDER BY dari sort string.
// "-created_at" → ORDER BY created_at DESC
// "name"        → ORDER BY _data->>'name' ASC
func BuildOrderClause(sort string) string {
	if sort == "" {
		return "created_at DESC"
	}
	if field, ok := strings.CutPrefix(sort, "-"); ok {
		if isTopLevelField(field) {
			return fmt.Sprintf("%s DESC", field)
		}
		return fmt.Sprintf("_data->>'%s' DESC", field)
	}
	if isTopLevelField(sort) {
		return fmt.Sprintf("%s ASC", sort)
	}
	return fmt.Sprintf("_data->>'%s' ASC", sort)
}

// isTopLevelField memeriksa apakah field adalah kolom tabel (bukan JSONB field).
func isTopLevelField(field string) bool {
	topLevel := map[string]bool{
		"id": true, "tenant_id": true, "company_id": true,
		"created_at": true, "updated_at": true,
		"_sync_status": true, "_sync_version": true,
	}
	return topLevel[field]
}
