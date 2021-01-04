// Package dmysql ...
package dmysql

import (
	"bytes"
	"fmt"
	"strings"
)

// Cond ...
type Cond struct {
	op    command
	field string
	val   interface{}
}

func gen(conds []*Cond, orderBy, groupBy string, exp bool) (sqlPattern string, args []interface{}) {
	if len(conds) == 0 {
		return
	}
	bf := &bytes.Buffer{}
	bf.WriteString("WHERE")
	for i, cond := range conds {
		switch cond.op {
		case commandEQ:
			bf.WriteString(" `" + cond.field + "` = ?")
			args = append(args, cond.val)
		case commandNE:
			bf.WriteString(" `" + cond.field + "` != ?")
			args = append(args, cond.val)
		case commandGT:
			bf.WriteString(" `" + cond.field + "` > ?")
			args = append(args, cond.val)
		case commandGTE:
			bf.WriteString(" `" + cond.field + "` >= ?")
			args = append(args, cond.val)
		case commandLT:
			bf.WriteString(" `" + cond.field + "` < ?")
			args = append(args, cond.val)
		case commandLTE:
			bf.WriteString(" `" + cond.field + "` <= ?")
			args = append(args, cond.val)
		case commandLIKE:
			bf.WriteString(" `" + cond.field + "` LIKE " + fmt.Sprint(cond.val))
		case commandNLIKE:
			bf.WriteString(" `" + cond.field + "` NOT LIKE " + fmt.Sprint(cond.val))
		case commandIN:
			vals, ok := cond.val.([]interface{})
			if ok {
				bf.WriteString(" `" + cond.field + "` IN (")
				if len(vals) > 1 {
					bf.WriteString(strings.Repeat("?,", len(vals)-1))
				}
				bf.WriteString("?)")
				args = append(args, vals...)
			} else {
				bf.WriteString(" `" + cond.field + "` IN " + fmt.Sprintln(cond.val))
			}
		case commandNIN:
			vals, ok := cond.val.([]interface{})
			if ok {
				bf.WriteString(" `" + cond.field + "` NOT IN (")
				if len(vals) > 1 {
					bf.WriteString(strings.Repeat("?,", len(vals)-1))
				}
				bf.WriteString("?)")
				args = append(args, vals...)
			} else {
				bf.WriteString(" `" + cond.field + "` NOT IN " + fmt.Sprintln(cond.val))
			}
		}
		if i != len(conds)-1 {
			if exp {
				bf.WriteString(" AND")
			} else {
				bf.WriteString(" OR")
			}
		}
	}
	if orderBy != "" {
		if strings.HasPrefix(orderBy, "-") {
			bf.WriteString(" ORDER BY `" + string([]byte(orderBy)[1:]) + "` DESC")
		} else if strings.HasPrefix(orderBy, "+") {
			bf.WriteString(" ORDER BY `" + string([]byte(orderBy)[1:]) + "`")
		} else {
			bf.WriteString(" ORDER BY `" + orderBy + "`")
		}
	}
	if groupBy != "" {
		bf.WriteString(" GROUP BY `" + groupBy + "`")
	}
	sqlPattern = bf.String()
	return
}

type command uint8

const (
	commandEQ command = iota
	commandNE
	commandGT
	commandGTE
	commandLT
	commandLTE
	commandIN
	commandNIN
	commandLIKE
	commandNLIKE
)

// Eq ...
func Eq(field string, val interface{}) *Cond { return &Cond{commandEQ, field, val} }

// NotEq ...
func NotEq(field string, val interface{}) *Cond { return &Cond{commandNE, field, val} }

// Gt ...
func Gt(field string, val interface{}) *Cond { return &Cond{commandGT, field, val} }

// Gte ...
func Gte(field string, val interface{}) *Cond { return &Cond{commandGTE, field, val} }

// Lt ...
func Lt(field string, val interface{}) *Cond { return &Cond{commandLT, field, val} }

// Lte ...
func Lte(field string, val interface{}) *Cond { return &Cond{commandLTE, field, val} }

// In ...
func In(field string, val interface{}) *Cond { return &Cond{commandIN, field, val} }

// NotIn ...
func NotIn(field string, val interface{}) *Cond { return &Cond{commandNIN, field, val} }

// Like ...
func Like(field string, val interface{}) *Cond { return &Cond{commandLIKE, field, val} }

// NotLike ...
func NotLike(field string, val interface{}) *Cond { return &Cond{commandNLIKE, field, val} }

// Or ...
func Or(conds []*Cond, orderBy, groupBy string) (sqlPattern string, args []interface{}) {
	return gen(conds, orderBy, groupBy, false)
}

// And ...
func And(conds []*Cond, orderBy, groupBy string) (sqlPattern string, args []interface{}) {
	return gen(conds, orderBy, groupBy, true)
}
