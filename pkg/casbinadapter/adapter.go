package casbinadapter

import (
	"context"
	"fmt"
	"strings"

	"github.com/Roisfaozi/go-clean-boilerplate/pkg/tx"
	"github.com/casbin/casbin/v3/model"
	"github.com/casbin/casbin/v3/persist"
)

type CasbinRule struct {
	ID    int64   `db:"id"`
	Ptype string  `db:"ptype"`
	V0    *string `db:"v0"`
	V1    *string `db:"v1"`
	V2    *string `db:"v2"`
	V3    *string `db:"v3"`
	V4    *string `db:"v4"`
	V5    *string `db:"v5"`
}

type SQLXCasbinAdapter struct {
	dbtx tx.DBTX
}

func NewSQLXCasbinAdapter(dbtx tx.DBTX) *SQLXCasbinAdapter {
	return &SQLXCasbinAdapter{dbtx: dbtx}
}

func (a *SQLXCasbinAdapter) LoadPolicy(m model.Model) error {
	ctx := context.Background()
	query := `SELECT id, ptype, v0, v1, v2, v3, v4, v5 FROM casbin_rule`
	var rules []CasbinRule
	if err := a.dbtx.SelectContext(ctx, &rules, query); err != nil {
		return err
	}

	for _, rule := range rules {
		loadRule(rule, m)
	}
	return nil
}

func loadRule(rule CasbinRule, m model.Model) {
	lineText := rule.Ptype
	vals := []*string{rule.V0, rule.V1, rule.V2, rule.V3, rule.V4, rule.V5}
	for len(vals) > 0 && (vals[len(vals)-1] == nil || *vals[len(vals)-1] == "") {
		vals = vals[:len(vals)-1]
	}

	policy := make([]string, 0, len(vals))
	for _, v := range vals {
		if v == nil {
			policy = append(policy, "")
			continue
		}
		policy = append(policy, *v)
	}
	if len(policy) > 0 {
		lineText += ", " + strings.Join(policy, ", ")
	}
	_ = persist.LoadPolicyLine(lineText, m)
}

func (a *SQLXCasbinAdapter) SavePolicy(m model.Model) error {
	ctx := context.Background()
	dropQuery := `DELETE FROM casbin_rule`
	if _, err := a.dbtx.ExecContext(ctx, dropQuery); err != nil {
		return err
	}

	for ptype, ast := range m["p"] {
		for _, rule := range ast.Policy {
			if err := a.saveRule(ctx, ptype, rule); err != nil {
				return err
			}
		}
	}

	for ptype, ast := range m["g"] {
		for _, rule := range ast.Policy {
			if err := a.saveRule(ctx, ptype, rule); err != nil {
				return err
			}
		}
	}

	return nil
}

func (a *SQLXCasbinAdapter) saveRule(ctx context.Context, ptype string, rule []string) error {
	cols := []string{"ptype"}
	args := []any{ptype}
	for i, val := range rule {
		cols = append(cols, fmt.Sprintf("v%d", i))
		args = append(args, val)
	}

	placeholders := make([]string, len(cols))
	for i := range placeholders {
		placeholders[i] = "?"
	}

	query := fmt.Sprintf(`INSERT INTO casbin_rule (%s) VALUES (%s)`,
		strings.Join(cols, ", "),
		strings.Join(placeholders, ", "),
	)

	_, err := a.dbtx.ExecContext(ctx, query, args...)
	return err
}

func (a *SQLXCasbinAdapter) AddPolicy(sec string, ptype string, rule []string) error {
	return a.saveRule(context.Background(), ptype, rule)
}

func (a *SQLXCasbinAdapter) AddPolicies(sec string, ptype string, rules [][]string) error {
	ctx := context.Background()
	for _, rule := range rules {
		if err := a.saveRule(ctx, ptype, rule); err != nil {
			return err
		}
	}
	return nil
}

func (a *SQLXCasbinAdapter) UpdatePolicy(sec string, ptype string, oldRule, newRule []string) error {
	ctx := context.Background()
	setClauses := make([]string, len(newRule))
	setArgs := make([]any, len(newRule))
	for i, val := range newRule {
		if i >= 6 {
			break
		}
		setClauses[i] = fmt.Sprintf("v%d = ?", i)
		setArgs[i] = val
	}

	whereClauses := []string{"ptype = ?"}
	whereArgs := []any{ptype}
	for i, val := range oldRule {
		if i >= 6 {
			break
		}
		whereClauses = append(whereClauses, fmt.Sprintf("v%d = ?", i))
		whereArgs = append(whereArgs, val)
	}

	query := fmt.Sprintf(`UPDATE casbin_rule SET %s WHERE %s`,
		strings.Join(setClauses, ", "),
		strings.Join(whereClauses, " AND "),
	)
	args := append(setArgs, whereArgs...)
	_, err := a.dbtx.ExecContext(ctx, query, args...)
	return err
}

func (a *SQLXCasbinAdapter) UpdatePolicies(sec string, ptype string, oldRules, newRules [][]string) error {
	if len(oldRules) != len(newRules) {
		return fmt.Errorf("the length of oldRules should be equal to the length of newRules, but got %d and %d", len(oldRules), len(newRules))
	}
	for i := range oldRules {
		if err := a.UpdatePolicy(sec, ptype, oldRules[i], newRules[i]); err != nil {
			return err
		}
	}
	return nil
}

func (a *SQLXCasbinAdapter) UpdateFilteredPolicies(sec string, ptype string, newRules [][]string, fieldIndex int, fieldValues ...string) ([][]string, error) {
	ctx := context.Background()
	whereClauses := []string{"ptype = ?"}
	args := []any{ptype}

	for i, val := range fieldValues {
		if val != "" {
			whereClauses = append(whereClauses, fmt.Sprintf("v%d = ?", fieldIndex+i))
			args = append(args, val)
		}
	}

	oldRules, err := a.queryRules(ctx, strings.Join(whereClauses, " AND "), args...)
	if err != nil {
		return nil, err
	}

	deleteQuery := fmt.Sprintf(`DELETE FROM casbin_rule WHERE %s`, strings.Join(whereClauses, " AND "))
	if _, err := a.dbtx.ExecContext(ctx, deleteQuery, args...); err != nil {
		return nil, err
	}

	for _, rule := range newRules {
		if err := a.saveRule(ctx, ptype, rule); err != nil {
			return oldRules, err
		}
	}
	return oldRules, nil
}

func (a *SQLXCasbinAdapter) queryRules(ctx context.Context, where string, args ...any) ([][]string, error) {
	query := fmt.Sprintf(`SELECT v0, v1, v2, v3, v4, v5 FROM casbin_rule WHERE %s`, where)
	rows, err := a.dbtx.QueryxContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = rows.Close()
	}()

	var result [][]string
	for rows.Next() {
		var vals [6]*string
		if err := rows.Scan(&vals[0], &vals[1], &vals[2], &vals[3], &vals[4], &vals[5]); err != nil {
			return nil, err
		}
		rule := make([]string, 6)
		for i := range vals {
			if vals[i] != nil {
				rule[i] = *vals[i]
			}
		}
		for len(rule) > 0 && rule[len(rule)-1] == "" {
			rule = rule[:len(rule)-1]
		}
		result = append(result, rule)
	}
	return result, rows.Err()
}

func (a *SQLXCasbinAdapter) RemovePolicy(sec string, ptype string, rule []string) error {
	ctx := context.Background()
	whereClauses := []string{"ptype = ?"}
	args := []any{ptype}
	for i, val := range rule {
		whereClauses = append(whereClauses, fmt.Sprintf("v%d = ?", i))
		args = append(args, val)
	}

	query := fmt.Sprintf(`DELETE FROM casbin_rule WHERE %s`, strings.Join(whereClauses, " AND "))
	_, err := a.dbtx.ExecContext(ctx, query, args...)
	return err
}

func (a *SQLXCasbinAdapter) RemovePolicies(sec string, ptype string, rules [][]string) error {
	for _, rule := range rules {
		if err := a.RemovePolicy(sec, ptype, rule); err != nil {
			return err
		}
	}
	return nil
}

func (a *SQLXCasbinAdapter) RemoveFilteredPolicy(sec string, ptype string, fieldIndex int, fieldValues ...string) error {
	ctx := context.Background()
	whereClauses := []string{"ptype = ?"}
	args := []any{ptype}

	for i, val := range fieldValues {
		if val != "" {
			whereClauses = append(whereClauses, fmt.Sprintf("v%d = ?", fieldIndex+i))
			args = append(args, val)
		}
	}

	query := fmt.Sprintf(`DELETE FROM casbin_rule WHERE %s`, strings.Join(whereClauses, " AND "))
	_, err := a.dbtx.ExecContext(ctx, query, args...)
	return err
}
