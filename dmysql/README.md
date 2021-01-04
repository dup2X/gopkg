# mysql
>mysql low level api

## Useage

 step 1:
    ```
    // 新建一个mysql管理器
	mgr, err := New(hosts, usr, passwd, db, charset,
		WithMaxConnSize(16), // 最大连接数 
		WithDebug(true), // 启用debug
		WithDialTimeout(time.Second*1), // 连接超时时间
		WithReadTimeout(time.Second*2), // 读超时
		WithWriteTimeout(time.Second*2), // 写超时
		WithAutoCommit(true), // 启用自动提交
		WithPoolSize(4)) // 连接池大小
    ```
step 2:
    ```
    conn,err := mgr.Get() // 获取一个连接
    defer mgr.Put(conn) // 将连接归还到连接池
    // conn.Xxxx 使用连接操作mysql
    ```

## API
  - Select
    Select(table string, fields []string, condPattern string, condArgs ...interface{})
    table:表名
    fields:指定列名，[]string{"*"}代表所有
    condPattern: where条件，使用占位符表达式
    condArgs: condPattern中对应的占位符的值
    ```
	err = conn.Select(table, []string{"name", "age"}, "where name = ?", "update1")
	row, err := conn.FetchRowMap()
    ```
  - Insert
	Insert(table string,params map[string]interface{})
    params: map[field] = value
    ```
	id, err = conn.Insert(table, map[string]interface{}{"name": "ja"})
    // id 为lastInsertID
    ```
  - Update
    Update(table string, updator map[string]interface{}, condPattern string,condArgs ...interface{}) (affect int64, err error) 
    updator: 更新的字段和对应的新值
    affect: 影响的行数
    ```
	_, err = conn.Update(table, map[string]interface{}{"name": "update1", "age": 222}, "where name = ?", "update")
    ```
  - Delete
    ```
    ```
  - Query
    ```
    ```
  - Execute
    ```
    ```
  - Upsert

## FAQ
1、防注入支持吗
你别自己拼SQL条件就行，底层有防注入实现，条件建议使用占位符
