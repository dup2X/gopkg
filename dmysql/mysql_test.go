package dmysql

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/dup2X/gopkg/logger"
)

var (
	createDatabaseSQL = "CREATE SCHEMA `test` DEFAULT CHARACTER SET utf8;"
	createTableSQL    = "CREATE TABLE `test`.`demo` (`id` INT NOT NULL AUTO_INCREMENT COMMENT 'ID',`name` VARCHAR(256) NOT NULL COMMENT 'NAME',PRIMARY KEY (`id`));"
	dropDatabaseSQL   = "DROP TABLE `demo`;"

	hosts   = []string{"127.0.0.1:3306", "127.0.0.1:3307"}
	usr     = "root"
	passwd  = "qwe123"
	db      = "test"
	charset = "utf8"
	table   = "test.demo"
	ctx     = context.Background()
)

func TestConnect(t *testing.T) {
	mgr, err := New(hosts[:1], usr, passwd, db, charset,
		WithMaxConnSize(16),
		WithDebug(true),
		WithDialTimeout(time.Second*1),
		WithReadTimeout(time.Second*2),
		WithWriteTimeout(time.Second*2),
		WithAutoCommit(true),
		WithPoolSize(4),
		WithLogger(logger.GetLoggers()),
	)
	if err != nil {
		t.Fatal(err.Error())
	}
	c1, err := mgr.Get()
	if err != nil {
		t.Fatal(err.Error())
	}
	println(time.Now().Unix())
	err = c1.Query(ctx, "select sleep(3)")
	if err == nil {
		t.Fatal(err.Error())
	} else {
		println("select sleep(3) timeout:", err.Error())
	}
	println(time.Now().Unix())
	c2, err := mgr.Get()
	if err != nil {
		t.Fatal(err.Error())
	}
	c3, err := mgr.Get()
	if err != nil {
		t.Fatal(err.Error())
	}
	c4, err := mgr.Get()
	if err != nil {
		t.Fatal(err.Error())
	}
	c5, err := mgr.Get()
	if err == nil || c5 != nil {
		t.Fatal(err.Error())
	}
	c1.Close()
	c2.Close()
	c3.Close()
	c4.Close()
}

func TestReconnect(t *testing.T) {
	mgr, err := New(hosts, usr, passwd, db, charset,
		WithMaxConnSize(8),
		WithDebug(true),
		WithDialTimeout(time.Second*1),
		WithReadTimeout(time.Second*2),
		WithWriteTimeout(time.Second*2),
		WithAutoCommit(true),
		WithKeepSilent(true),
		WithPoolSize(4))
	if err != nil {
		t.Fatal(err.Error())
	}
	c1, err := mgr.Get()
	if err != nil {
		t.Fatal(err.Error())
	}

	for i := 0; i < 3; i++ {
		time.Sleep(time.Millisecond * 20)
		err = c1.Select(ctx, table, []string{"*"}, "")
		if err != nil {
			println(i, err.Error())
			continue
		}
		_, err = c1.FetchAllMap(ctx)
		if err != nil {
			println(err.Error())
		}
	}
}

func initDatabase() {
	mgr, err := New(hosts, usr, passwd, db, charset, WithKeepSilent(true))
	if err != nil {
		println(err.Error())
		return
	}
	conn, err := mgr.Get()
	if err != nil {
		println(err.Error())
		return
	}
	conn.Execute(ctx, createTableSQL)
}

func destroyDatabase() {
	mgr, err := New(hosts, usr, passwd, db, charset, WithKeepSilent(true))
	if err != nil {
		println(err.Error())
		return
	}
	conn, err := mgr.Get()
	if err != nil {
		println(err.Error())
		return
	}
	conn.Execute(ctx, dropDatabaseSQL)
}

func TestInsert(t *testing.T) {
	mgr, err := New(hosts, usr, passwd, db, charset,
		WithMaxConnSize(16),
		WithDebug(true),
		WithAutoCommit(true),
		WithDialTimeout(time.Second*1),
		WithReadTimeout(time.Second*2),
		WithWriteTimeout(time.Second*2),
		WithKeepSilent(true),
		WithPoolSize(4))
	if err != nil {
		t.Fatal(err.Error())
	}
	conn, err := mgr.Get()
	if err != nil {
		t.Fatal(err.Error())
	}
	defer mgr.Put(conn)
	_, err = conn.Insert(ctx, table, map[string]interface{}{"name": "ja"})
	if err != nil {
		t.Fatal(err.Error())
	}
	_, err = conn.Delete(ctx, table, "where name = ?", "ja")
	if err != nil {
		t.Fatal(err.Error())
	}
}

func TestMultiInsert(t *testing.T) {
	mgr, err := New(hosts, usr, passwd, db, charset,
		WithMaxConnSize(16),
		WithDebug(true),
		WithDialTimeout(time.Second*1),
		WithReadTimeout(time.Second*2),
		WithWriteTimeout(time.Second*2),
		WithKeepSilent(true),
		WithPoolSize(4))
	if err != nil {
		t.Fatal(err.Error())
	}
	conn, err := mgr.Get()
	if err != nil {
		t.Fatal(err.Error())
	}
	defer mgr.Put(conn)
	_, err = conn.MultiInsert(ctx, table, []map[string]interface{}{
		{"name": "ja"},
		{"name": "ja"},
		{"name": "ja"},
	})
	if err != nil {
		t.Fatal(err.Error())
	}
	_, err = conn.Delete(ctx, table, "where name = ?", "ja")
	if err != nil {
		t.Fatal(err.Error())
	}
}

func TestUpsert(t *testing.T) {
	mgr, err := New(hosts, usr, passwd, db, charset,
		WithMaxConnSize(8),
		WithDebug(true),
		WithDialTimeout(time.Second*1),
		WithReadTimeout(time.Second*2),
		WithWriteTimeout(time.Second*2),
		WithKeepSilent(true),
		WithPoolSize(4))
	if err != nil {
		t.Fatal(err.Error())
	}
	conn, err := mgr.Get()
	if err != nil {
		t.Fatal(err.Error())
	}
	defer mgr.Put(conn)
	_, err = conn.Upsert(ctx, table, map[string]interface{}{"name": "upsert"}, []string{"name"})
	if err != nil {
		t.Fatal(err.Error())
	}
	conn.Delete(ctx, table, "where name=?", "upsert")
}

func TestUpdate(t *testing.T) {
	mgr, err := New(hosts, usr, passwd, db, charset, WithKeepSilent(true))
	if err != nil {
		t.Fatal(err.Error())
	}
	conn, err := mgr.Get()
	if err != nil {
		t.Fatal(err.Error())
	}
	defer mgr.Put(conn)
	_, err = conn.Insert(ctx, table, map[string]interface{}{"name": "update", "age": 1})
	if err != nil {
		t.Fatal(err.Error())
	}

	_, err = conn.Update(ctx, table, map[string]interface{}{"name": "update1", "age": 222}, "where name = ?", "update")
	if err != nil {
		t.Fatal(err.Error())
	}
	err = conn.Select(ctx, table, []string{"name", "age"}, "where name = ?", "update1")
	if err != nil {
		t.Fatal(err.Error())
	}
	row, err := conn.FetchRowMap(ctx)
	if err != nil || len(row) == 0 {
		t.Fatal(err.Error())
	}
	if row["name"] != "update1" && row["age"] != "222" {
		t.FailNow()
	}
	conn.Delete(ctx, table, "where name = ?", "update1")
}

func TestSelect(t *testing.T) {
	mgr, err := New(hosts, usr, passwd, db, charset, WithKeepSilent(true))
	if err != nil {
		t.Fatal(err.Error())
	}
	conn, err := mgr.Get()
	if err != nil {
		t.Fatal(err.Error())
	}
	defer mgr.Put(conn)
	_, err = conn.Insert(ctx, table, map[string]interface{}{"name": "q"})
	if err != nil {
		t.Fatal(err.Error())
	}
	err = conn.Select(ctx, table, []string{"id", "name"}, "WHERE name=?", "q")
	if err != nil {
		t.Fatal(err.Error())
	}
	defer conn.Delete(ctx, table, "WHERE name=?", "q")
	all, err := conn.FetchAllMap(ctx)
	if err != nil {
		t.Fatal(err.Error())
	}
	fmt.Printf("%v\n", all)
}

func TestTrans(t *testing.T) {
	mgr, err := New(hosts, usr, passwd, db, charset, WithKeepSilent(true))
	if err != nil {
		t.Fatal(err.Error())
	}
	conn, err := mgr.Get()
	if err != nil {
		t.Fatal(err.Error())
	}
	defer mgr.Put(conn)
	_, err = conn.Insert(ctx, table, map[string]interface{}{"name": "tr"})
	if err != nil {
		t.Fatal(err.Error())
	}
	err = conn.Begin(ctx)
	if err != nil {
		t.Fatal(err.Error())
	}
	sql := fmt.Sprintf("UPDATE %s SET name = 'tr1' WHERE name = ?", table)
	err = conn.Execute(ctx, sql, "tr")
	if err != nil {
		conn.RollBack(ctx)
		t.Fatal(err.Error())
	}
	err = conn.Commit(ctx)
	if err != nil {
		t.Fatal(err.Error())
	}
	conn.Delete(ctx, table, "WHERE name=?", "tr1")
}

func TestFetch(t *testing.T) {
	mgr, err := New(hosts, usr, passwd, db, charset, WithDebug(true), WithKeepSilent(true))
	if err != nil {
		t.Fatal(err.Error())
	}
	conn, err := mgr.Get()
	if err != nil {
		t.Fatal(err.Error())
	}
	defer mgr.Put(conn)

	err = conn.Select(ctx, table, []string{"*"}, "")
	if err != nil {
		t.Fatal(err.Error())
	}
	res, err := conn.FetchAll(ctx)
	if err != nil {
		t.Fatal(err.Error())
	}
	println("fetchAll:", len(res))

	err = conn.Select(ctx, table, []string{"*"}, "")
	if err != nil {
		t.Fatal(err.Error())
	}
	res1, err := conn.FetchRowMapInterface(ctx)
	if err != nil {
		t.Fatal(err.Error())
	}
	println("fetchAll:", len(res1))

	err = conn.Select(ctx, table, []string{"*"}, "where id < ?", 1)
	if err != nil {
		println(err.Error())
		t.Fatal(err.Error())
	}
	res2, err := conn.FetchRowMap(ctx)
	if err != nil {
		println(err.Error())
		t.Fatal(err.Error())
	}
	println("should nil", len(res2))
}

func BenchmarkQuery(b *testing.B) {
	mgr, err := New(hosts, usr, passwd, db, charset, WithKeepSilent(true))
	if err != nil {
		b.Fatal(err.Error())
	}
	conn, err := mgr.Get()
	if err != nil {
		b.Fatal(err.Error())
	}
	for i := 0; i < b.N; i++ {
		conn.Query(ctx, "select * from user")
		conn.FetchAll(ctx)
	}
	mgr.Put(conn)
}
