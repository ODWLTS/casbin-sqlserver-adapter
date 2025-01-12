// Copyright 2021 by ODW LTS. All Rights Reserved.
// Copyright 2020 by Blank-Xu. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//
// This file file has been modified by ODW LTS to improve SQL Server support.

package casbinsqlserveradapter

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"strconv"

	"github.com/casbin/casbin/v2/model"
	"github.com/casbin/casbin/v2/persist"
	"github.com/jmoiron/sqlx"
)

// dafaultSchemaName if schemaName == "", the Adapter will use the default schema.
const defaultSchemaName = "dbo"

// defaultTableName  if tableName == "", the Adapter will use this default table name.
const defaultTableName = "casbin_rule"

// maxParamLength  .
const maxParamLength = 7

// CasbinRule  defines the casbin rule model.
// It used for save or load policy lines from sqlx connected database.
type CasbinRule struct {
	PType string `db:"p_type"`
	V0    string `db:"v0"`
	V1    string `db:"v1"`
	V2    string `db:"v2"`
	V3    string `db:"v3"`
	V4    string `db:"v4"`
	V5    string `db:"v5"`
}

// Adapter  define the sqlx adapter for Casbin.
// It can load policy lines or save policy lines from sqlx connected database.
type Adapter struct {
	db  *sqlx.DB
	ctx context.Context

	schemaName string
	tableName  string

	isFiltered bool

	sqlCreateTable   string
	sqlTruncateTable string
	sqlIsTableExist  string
	sqlInsertRow     string
	sqlUpdateRow     string
	sqlDeleteAll     string
	sqlDeleteRow     string
	sqlDeleteByArgs  string
	sqlSelectAll     string
	sqlSelectWhere   string
}

// Filter  defines the filtering rules for a FilteredAdapter's policy.
// Empty values are ignored, but all others must match the filter.
type Filter struct {
	PType []string
	V0    []string
	V1    []string
	V2    []string
	V3    []string
	V4    []string
	V5    []string
}

// V0IncludeV2ExcludeFilter defines the filtering include and exclude rules for
// a FilteredAdapter's policy.
type IncludeExcludeFilter struct {
	PTypeInclude []string
	PTypeExclude []string
	V0Include    []string
	V0Exclude    []string
	V1Include    []string
	V1Exclude    []string
	V2Include    []string
	V2Exclude    []string
	V3Include    []string
	V3Exclude    []string
	V4Include    []string
	V4Exclude    []string
	V5Include    []string
	V5Exclude    []string
}

// NewAdapter  the constructor for Adapter.
// db should connected to database and controlled by user.
// If tableName == "", the Adapter will automatically create a table named "casbin_rule".
func NewAdapter(db *sqlx.DB, schemaName, tableName string) (*Adapter, error) {
	return NewAdapterContext(context.Background(), db, schemaName, tableName)
}

// NewAdapterContext  the constructor for Adapter.
// db should connected to database and controlled by user.
// If tableName == "", the Adapter will automatically create a table named "casbin_rule".
func NewAdapterContext(ctx context.Context, db *sqlx.DB, schemaName, tableName string) (*Adapter, error) {
	if db == nil {
		return nil, errors.New("db is nil")
	}

	// check db connecting
	err := db.PingContext(ctx)
	if err != nil {
		return nil, err
	}

	switch db.DriverName() {
	case "oci8", "ora", "goracle":
		return nil, errors.New("sqlxadapter: please checkout 'oracle' branch")
	}

	if schemaName == "" {
		schemaName = defaultSchemaName
	}

	if tableName == "" {
		tableName = defaultTableName
	}

	adapter := Adapter{
		db:         db,
		ctx:        ctx,
		schemaName: schemaName,
		tableName:  tableName,
	}

	// generate different databases sql
	adapter.genSQL()

	if !adapter.isTableExist() {
		if err = adapter.createTable(); err != nil {
			return nil, err
		}
	}

	return &adapter, nil
}

// genSQL  generate sql based on db driver name.
func (p *Adapter) genSQL() {
	p.sqlCreateTable = fmt.Sprintf(sqlCreateTable, p.schemaName, p.tableName)
	p.sqlTruncateTable = fmt.Sprintf(sqlTruncateTable, p.schemaName, p.tableName)

	p.sqlIsTableExist = fmt.Sprintf(sqlIsTableExist, p.schemaName, p.tableName)

	p.sqlInsertRow = fmt.Sprintf(sqlInsertRow, p.schemaName, p.tableName)
	p.sqlUpdateRow = fmt.Sprintf(sqlUpdateRow, p.schemaName, p.tableName)
	p.sqlDeleteAll = fmt.Sprintf(sqlDeleteAll, p.schemaName, p.tableName)
	p.sqlDeleteRow = fmt.Sprintf(sqlDeleteRow, p.schemaName, p.tableName)
	p.sqlDeleteByArgs = fmt.Sprintf(sqlDeleteByArgs, p.schemaName, p.tableName)

	p.sqlSelectAll = fmt.Sprintf(sqlSelectAll, p.schemaName, p.tableName)
	p.sqlSelectWhere = fmt.Sprintf(sqlSelectWhere, p.schemaName, p.tableName)
}

// createTable  create a not exists table.
func (p *Adapter) createTable() error {
	_, err := p.db.ExecContext(p.ctx, p.sqlCreateTable)

	return err
}

// truncateTable  clear the table.
func (p *Adapter) truncateTable() error {
	_, err := p.db.ExecContext(p.ctx, p.sqlTruncateTable)

	return err
}

// deleteAll  clear the table.
func (p *Adapter) deleteAll() error {
	_, err := p.db.ExecContext(p.ctx, p.sqlDeleteAll)

	return err
}

// isTableExist  check the table exists.
func (p *Adapter) isTableExist() bool {
	_, err := p.db.ExecContext(p.ctx, p.sqlIsTableExist)

	return err == nil
}

// deleteRows  delete eligible data.
func (p *Adapter) deleteRows(query string, args ...interface{}) error {
	query = p.db.Rebind(query)

	_, err := p.db.ExecContext(p.ctx, query, args...)

	return err
}

// truncateAndInsertRows  clear table and insert new rows.
func (p *Adapter) truncateAndInsertRows(rules [][]interface{}) error {
	if err := p.truncateTable(); err != nil {
		return err
	}
	return p.execTxSqlRows(p.sqlInsertRow, rules)
}

// deleteAllAndInsertRows  clear table and insert new rows.
func (p *Adapter) deleteAllAndInsertRows(rules [][]interface{}) error {
	if err := p.deleteAll(); err != nil {
		return err
	}
	return p.execTxSqlRows(p.sqlInsertRow, rules)
}

// execTxSqlRows  exec sql rows.
func (p *Adapter) execTxSqlRows(query string, rules [][]interface{}) (err error) {
	tx, err := p.db.BeginTx(p.ctx, nil)
	if err != nil {
		return
	}

	var action string

	stmt, err := tx.PrepareContext(p.ctx, query)
	if err != nil {
		action = "prepare context"
		goto ROLLBACK
	}

	for _, rule := range rules {
		if _, err = stmt.ExecContext(p.ctx, rule...); err != nil {
			action = "stmt exec"
			goto ROLLBACK
		}
	}

	if err = stmt.Close(); err != nil {
		action = "stmt close"
		goto ROLLBACK
	}

	if err = tx.Commit(); err != nil {
		action = "commit"
		goto ROLLBACK
	}

	return

ROLLBACK:

	if err1 := tx.Rollback(); err1 != nil {
		err = fmt.Errorf("%s err: %v, rollback err: %v", action, err, err1)
	}

	return
}

// selectRows  select eligible data by args from the table.
func (p *Adapter) selectRows(query string, args ...interface{}) ([]*CasbinRule, error) {
	// make a slice with capacity
	lines := make([]*CasbinRule, 0, 64)

	if len(args) == 0 {
		return lines, p.db.SelectContext(p.ctx, &lines, query)
	}

	query = p.db.Rebind(query)

	return lines, p.db.SelectContext(p.ctx, &lines, query, args...)
}

// selectWhereIn  select eligible data by filter from the table.
func (p *Adapter) selectWhereIn(filter *Filter) (lines []*CasbinRule, err error) {
	var sqlBuf bytes.Buffer

	sqlBuf.Grow(64)
	sqlBuf.WriteString(p.sqlSelectWhere)

	args := make([]interface{}, 0, 4)

	hasInCond := false

	for _, col := range [maxParamLength]struct {
		name string
		arg  []string
	}{
		{"p_type", filter.PType},
		{"v0", filter.V0},
		{"v1", filter.V1},
		{"v2", filter.V2},
		{"v3", filter.V3},
		{"v4", filter.V4},
		{"v5", filter.V5},
	} {
		l := len(col.arg)
		if l == 0 {
			continue
		}

		switch sqlBuf.Bytes()[sqlBuf.Len()-1] {
		case '?', ')':
			sqlBuf.WriteString(" AND ")
		}

		sqlBuf.WriteString(col.name)

		if l == 1 {
			sqlBuf.WriteString("=?")
			args = append(args, col.arg[0])
		} else {
			sqlBuf.WriteString(" IN (?)")
			args = append(args, col.arg)

			hasInCond = true
		}
	}

	var query string

	if hasInCond {
		if query, args, err = sqlx.In(sqlBuf.String(), args...); err != nil {
			return
		}
	} else {
		query = sqlBuf.String()
	}

	return p.selectRows(query, args...)
}

// selectWhereInNotIn select eligible data by filter from the table.
func (p *Adapter) selectWhereInNotIn(filter *IncludeExcludeFilter) (lines []*CasbinRule, err error) {
	var sqlBuf bytes.Buffer

	sqlBuf.Grow(64)
	sqlBuf.WriteString(p.sqlSelectWhere)

	args := make([]interface{}, 0, 4)

	hasInCond := false

	for _, col := range [maxParamLength * 2]struct {
		name    string
		include bool
		arg     []string
	}{
		{"p_type", true, filter.PTypeInclude},
		{"v0", true, filter.V0Include},
		{"v1", true, filter.V1Include},
		{"v2", true, filter.V2Include},
		{"v3", true, filter.V3Include},
		{"v4", true, filter.V4Include},
		{"v5", true, filter.V5Include},
		{"p_type", false, filter.PTypeExclude},
		{"v0", false, filter.V0Exclude},
		{"v1", false, filter.V1Exclude},
		{"v2", false, filter.V2Exclude},
		{"v3", false, filter.V3Exclude},
		{"v4", false, filter.V4Exclude},
		{"v5", false, filter.V5Exclude},
	} {
		l := len(col.arg)
		if l == 0 {
			continue
		}

		switch sqlBuf.Bytes()[sqlBuf.Len()-1] {
		case '?', ')':
			sqlBuf.WriteString(" AND ")
		}

		sqlBuf.WriteString(col.name)

		if col.include {
			if l == 1 {
				sqlBuf.WriteString(" = ?")
				args = append(args, col.arg[0])
			} else {
				sqlBuf.WriteString(" IN (?)")
				args = append(args, col.arg)

				hasInCond = true
			}
		} else {
			if l == 1 {
				sqlBuf.WriteString(" != ?")
				args = append(args, col.arg[0])
			} else {
				sqlBuf.WriteString(" NOT IN (?)")
				args = append(args, col.arg)

				hasInCond = true
			}
		}

	}

	var query string

	if hasInCond {
		if query, args, err = sqlx.In(sqlBuf.String(), args...); err != nil {
			return
		}
	} else {
		query = sqlBuf.String()
	}

	return p.selectRows(query, args...)
}

// LoadPolicy  load all policy rules from the storage.
func (p *Adapter) LoadPolicy(model model.Model) error {
	lines, err := p.selectRows(p.sqlSelectAll)
	if err != nil {
		return err
	}

	for _, line := range lines {
		p.loadPolicyLine(line, model)
	}

	return nil
}

// SavePolicy  save policy rules to the storage.
func (p *Adapter) SavePolicy(model model.Model) error {
	args := make([][]interface{}, 0, 64)

	for ptype, ast := range model["p"] {
		for _, rule := range ast.Policy {
			arg := p.genArgs(ptype, rule)
			args = append(args, arg)
		}
	}

	for ptype, ast := range model["g"] {
		for _, rule := range ast.Policy {
			arg := p.genArgs(ptype, rule)
			args = append(args, arg)
		}
	}

	return p.deleteAllAndInsertRows(args)
}

// AddPolicy  add one policy rule to the storage.
func (p *Adapter) AddPolicy(sec string, ptype string, rule []string) error {
	args := p.genArgs(ptype, rule)
	_, err := p.db.ExecContext(p.ctx, p.sqlInsertRow, args...)

	return err
}

// AddPolicies  add multiple policy rules to the storage.
func (p *Adapter) AddPolicies(sec string, ptype string, rules [][]string) error {
	args := make([][]interface{}, 0, 8)

	for _, rule := range rules {
		arg := p.genArgs(ptype, rule)
		args = append(args, arg)
	}

	return p.execTxSqlRows(p.sqlInsertRow, args)
}

// RemovePolicy  remove policy rules from the storage.
func (p *Adapter) RemovePolicy(sec string, ptype string, rule []string) error {
	var sqlBuf bytes.Buffer

	sqlBuf.Grow(64)
	sqlBuf.WriteString(p.sqlDeleteByArgs)

	args := make([]interface{}, 0, 4)
	args = append(args, ptype)

	for idx, arg := range rule {
		if arg != "" {
			sqlBuf.WriteString(" AND v")
			sqlBuf.WriteString(strconv.Itoa(idx))
			sqlBuf.WriteString("=?")

			args = append(args, arg)
		}
	}

	return p.deleteRows(sqlBuf.String(), args...)
}

// RemoveFilteredPolicy  remove policy rules that match the filter from the storage.
func (p *Adapter) RemoveFilteredPolicy(sec string, ptype string, fieldIndex int, fieldValues ...string) error {
	var sqlBuf bytes.Buffer

	sqlBuf.Grow(64)
	sqlBuf.WriteString(p.sqlDeleteByArgs)

	args := make([]interface{}, 0, 4)
	args = append(args, ptype)

	var value string

	l := fieldIndex + len(fieldValues)

	for idx := 0; idx < 6; idx++ {
		if fieldIndex <= idx && idx < l {
			value = fieldValues[idx-fieldIndex]

			if value != "" {
				sqlBuf.WriteString(" AND v")
				sqlBuf.WriteString(strconv.Itoa(idx))
				sqlBuf.WriteString("=?")

				args = append(args, value)
			}
		}
	}

	return p.deleteRows(sqlBuf.String(), args...)
}

// RemovePolicies  remove policy rules.
func (p *Adapter) RemovePolicies(sec string, ptype string, rules [][]string) (err error) {
	args := make([][]interface{}, 0, 8)

	for _, rule := range rules {
		arg := p.genArgs(ptype, rule)
		args = append(args, arg)
	}

	return p.execTxSqlRows(p.sqlDeleteRow, args)
}

// LoadFilteredPolicy  load policy rules that match the filter.
// filterPtr must be a pointer.
func (p *Adapter) LoadFilteredPolicy(model model.Model, filterPtr interface{}) error {
	if filterPtr == nil {
		return p.LoadPolicy(model)
	}

	filter, ok := filterPtr.(*Filter)
	if ok {
		lines, err := p.selectWhereIn(filter)
		if err != nil {
			return err
		}

		for _, line := range lines {
			p.loadPolicyLine(line, model)
		}

		p.isFiltered = true

		return nil
	} else {
		filterIncludeExclude, ok := filterPtr.(*IncludeExcludeFilter)
		if !ok {
			return errors.New("invalid filter type")
		}

		lines, err := p.selectWhereInNotIn(filterIncludeExclude)
		if err != nil {
			return err
		}

		for _, line := range lines {
			p.loadPolicyLine(line, model)
		}

		p.isFiltered = true

		return nil
	}
}

// IsFiltered  returns true if the loaded policy rules has been filtered.
func (p *Adapter) IsFiltered() bool {
	return p.isFiltered
}

// UpdatePolicy update a policy rule from storage.
// This is part of the Auto-Save feature.
func (p *Adapter) UpdatePolicy(sec, ptype string, oldRule, newPolicy []string) error {
	oldArg := p.genArgs(ptype, oldRule)
	newArg := p.genArgs(ptype, newPolicy)

	_, err := p.db.ExecContext(p.ctx, p.sqlUpdateRow, append(newArg, oldArg...)...)

	return err
}

// UpdatePolicies updates policy rules to storage.
func (p *Adapter) UpdatePolicies(sec, ptype string, oldRules, newRules [][]string) (err error) {
	if len(oldRules) != len(newRules) {
		return errors.New("old rules size not equal to new rules size")
	}

	args := make([][]interface{}, 0, 16)

	for idx := range oldRules {
		oldArg := p.genArgs(ptype, oldRules[idx])
		newArg := p.genArgs(ptype, newRules[idx])
		args = append(args, append(newArg, oldArg...))
	}

	return p.execTxSqlRows(p.sqlUpdateRow, args)
}

// UpdateFilteredPolicies deletes old rules and adds new rules.
func (p *Adapter) UpdateFilteredPolicies(sec, ptype string, newPolicies [][]string, fieldIndex int, fieldValues ...string) (oldPolicies [][]string, err error) {
	var value string

	var whereBuf bytes.Buffer
	whereBuf.Grow(32)

	l := fieldIndex + len(fieldValues)

	whereArgs := make([]interface{}, 0, 4)
	whereArgs = append(whereArgs, ptype)

	for idx := 0; idx < 6; idx++ {
		if fieldIndex <= idx && idx < l {
			value = fieldValues[idx-fieldIndex]

			if value != "" {
				whereBuf.WriteString(" AND v")
				whereBuf.WriteString(strconv.Itoa(idx))
				whereBuf.WriteString("=?")

				whereArgs = append(whereArgs, value)
			}
		}
	}

	var selectBuf bytes.Buffer
	selectBuf.Grow(64)
	selectBuf.WriteString(p.sqlSelectWhere)
	selectBuf.WriteString("p_type=?")
	selectBuf.Write(whereBuf.Bytes())

	var oldRows []*CasbinRule
	value = p.db.Rebind(selectBuf.String())
	oldRows, err = p.selectRows(value, whereArgs...)
	if err != nil {
		return
	}

	var deleteBuf bytes.Buffer
	deleteBuf.Grow(64)
	deleteBuf.WriteString(p.sqlDeleteByArgs)
	deleteBuf.Write(whereBuf.Bytes())

	var tx *sqlx.Tx
	tx, err = p.db.BeginTxx(p.ctx, nil)
	if err != nil {
		return
	}

	var (
		stmt   *sqlx.Stmt
		action string
	)
	value = p.db.Rebind(deleteBuf.String())
	if _, err = tx.ExecContext(p.ctx, value, whereArgs...); err != nil {
		action = "delete old policies"
		if err1 := tx.Rollback(); err1 != nil {
			err = fmt.Errorf("%s err: %v, rollback err: %v", action, err, err1)
			return
		}
	}

	stmt, err = tx.PreparexContext(p.ctx, p.sqlInsertRow)
	if err != nil {
		action = "preparex context"
		if err1 := tx.Rollback(); err1 != nil {
			err = fmt.Errorf("%s err: %v, rollback err: %v", action, err, err1)
			return
		}
	}

	for _, policy := range newPolicies {
		arg := p.genArgs(ptype, policy)
		if _, err = stmt.ExecContext(p.ctx, arg...); err != nil {
			action = "stmt exec context"
			if err1 := tx.Rollback(); err1 != nil {
				err = fmt.Errorf("%s err: %v, rollback err: %v", action, err, err1)
				return
			}
		}
	}

	if err = stmt.Close(); err != nil {
		action = "stmt close"
		if err1 := tx.Rollback(); err1 != nil {
			err = fmt.Errorf("%s err: %v, rollback err: %v", action, err, err1)
			return
		}
	}

	if err = tx.Commit(); err != nil {
		action = "commit"
		if err1 := tx.Rollback(); err1 != nil {
			err = fmt.Errorf("%s err: %v, rollback err: %v", action, err, err1)
			return
		}
	}

	oldPolicies = make([][]string, 0, len(oldRows))
	for _, rule := range oldRows {
		oldPolicies = append(oldPolicies, []string{rule.PType, rule.V0, rule.V1, rule.V2, rule.V3, rule.V4, rule.V5})
	}

	return
}

// loadPolicyLine  load a policy line to model.
func (Adapter) loadPolicyLine(line *CasbinRule, model model.Model) {
	if line == nil {
		return
	}

	var lineBuf bytes.Buffer

	lineBuf.Grow(64)
	lineBuf.WriteString(line.PType)

	args := [6]string{line.V0, line.V1, line.V2, line.V3, line.V4, line.V5}
	for _, arg := range args {
		if arg != "" {
			lineBuf.WriteByte(',')
			lineBuf.WriteString(arg)
		}
	}

	persist.LoadPolicyLine(lineBuf.String(), model)
}

// genArgs  generate args from ptype and rule.
func (Adapter) genArgs(ptype string, rule []string) []interface{} {
	l := len(rule)

	args := make([]interface{}, maxParamLength)
	args[0] = ptype

	for idx := 0; idx < l; idx++ {
		args[idx+1] = rule[idx]
	}

	for idx := l + 1; idx < maxParamLength; idx++ {
		args[idx] = ""
	}

	return args
}
