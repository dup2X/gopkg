package dmysql

import (
	"fmt"
	"testing"
)

func TestCond(t *testing.T) {
	conds := []*Cond{
		Eq("name", "james"),
		NotEq("age", 1),
		Lt("level", 21),
		Lte("le", 2),
		Gt("l", 120),
		Gte("m", 92),
		In("state", []interface{}{1, 2, 3}),
		NotIn("type", []interface{}{101, 102}),
		Like("atr", "'%xx'"),
		NotLike("atr1", "'%xx%'"),
	}
	sql, args := And(conds, "-id", "type")
	fmt.Println(sql)
	fmt.Printf("%v\n", args)

	sql, args = And(conds, "+id", "type")
	fmt.Println(sql)
	fmt.Printf("%v\n", args)

	conds = append(conds, In("gate", "(select gate from table)"))
	sql, args = And(conds, "id", "type")
	fmt.Println(sql)
	fmt.Printf("%v\n", args)

	sql, args = Or(conds, "id", "type")
	fmt.Println(sql)
	fmt.Printf("%v\n", args)
}
