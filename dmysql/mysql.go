// Package dmysql ...
package dmysql

import (
	"bytes"
	"context"
	"database/sql"
	"errors"
	"fmt"
	"io"
	"net"
	"strings"
	"sync"
	"time"

	dctx "github.com/dup2X/gopkg/context"
	"github.com/dup2X/gopkg/discovery"
	"github.com/dup2X/gopkg/elapsed"

	_ "github.com/go-sql-driver/mysql" // call init to register mysql driver
)

const (
	mysqlDriver     = "mysql"
	defaultCharset  = "utf8"
	defaultPoolSize = 8

	matchAllMask = "*"
)

var (
	// ErrEmptyValues nil val
	ErrEmptyValues = errors.New("values is nil")
	// ErrEmptyTable nil table name
	ErrEmptyTable = errors.New("table name is nil")
	// ErrEmptyCondition cond is nil
	ErrEmptyCondition = errors.New("where condition is nil")
	// ErrMissMatchRow not found any row
	ErrMissMatchRow = errors.New("missed match")
	// ErrNilTransaction transaction is nil
	ErrNilTransaction = errors.New("transaction is nil")
	// ErrUnsupportedMode unsupported AcquireConnMode
	ErrUnsupportedMode = errors.New("unsupported AcquireConnMode")
	// ErrEmptyConnPool empty connection pool
	ErrEmptyConnPool = errors.New("empty connection pool")
	// ErrAcquiredConnTimeout acquire connection timed out
	ErrAcquiredConnTimeout = errors.New("acquire connection timed out")
	// ErrNotOpened close or nil db
	ErrNotOpened        = errors.New("close or nil db")
	errNotSameFieldData = errors.New("batch data has different num of field")
)

// AcquireConnMode ...
type AcquireConnMode uint8

const (
	// AcquireConnModeUnblock return conn if there is conn or return nil if there is no free
	AcquireConnModeUnblock AcquireConnMode = iota
	// AcquireConnModeTimeout wait for free conn util timeout
	AcquireConnModeTimeout
	// AcquireConnModeBlock wait for free conn forever
	AcquireConnModeBlock
)

// Manager Mysql conn manager
type Manager struct {
	pool         chan *MySQL
	opt          *option
	Connected    int
	balancer     discovery.Balancer
	disfBalancer discovery.Balancer

	mu         sync.Mutex
	activeConn int64
}

// New return manager with some params and options
func New(hosts []string, usr, passwd, db, charset string, opts ...Option) (*Manager, error) {
	var err error
	copt := &connectionOption{
		usr:     usr,
		passwd:  passwd,
		dbname:  db,
		charset: charset,
	}
	opt := &option{copt: copt}
	for _, o := range opts {
		o(opt)
	}
	copt.log = opt.log
	if opt.mode == AcquireConnModeTimeout || opt.waitTimeout == 0 {
		opt.waitTimeout = time.Millisecond * 50
	}

	if opt.poolSize == 0 {
		opt.poolSize = defaultPoolSize
	}
	mgr := &Manager{
		pool:     make(chan *MySQL, opt.poolSize),
		opt:      opt,
		balancer: discovery.NewWithHosts(hosts),
	}
	cnt, err := mgr.initPool()
	mgr.Connected = cnt
	if mgr.opt.keepSilent && mgr.Connected > 0 {
		return mgr, nil
	}
	return mgr, err
}

func (mgr *Manager) initPool() (usable int, err error) {
	mgr.pool = make(chan *MySQL, mgr.opt.poolSize)
	for i := 0; i < mgr.opt.poolSize; i++ {
		conn, err := mgr.newDB()
		if err != nil {
			if mgr.opt.keepSilent {
				continue
			}
			return usable, err
		}
		err = conn.Ping()
		if err != nil {
			if mgr.opt.keepSilent {
				continue
			}
			return usable, err
		}
		mgr.Put(conn)
		usable++
	}
	return usable, nil
}

func (mgr *Manager) newDB() (*MySQL, error) {
	var (
		err  error
		addr string
	)
	if mgr.disfBalancer != nil {
		addr, err = mgr.disfBalancer.Get()
	}
	if addr == "" || err != nil {
		addr, err = mgr.balancer.Get()
	}
	if err != nil {
		return nil, err
	}
	hp, err := newAddress(addr)
	if err != nil {
		return nil, err
	}
	db := newMySQL(hp, mgr.opt.copt)
	err = db.connect()
	if err != nil {
		return nil, err
	}
	return db, nil
}

// Put put back conn into pool. Close it if pool is full
func (mgr *Manager) Put(db *MySQL) {
	// TODO connect other when one is down
	// if db.db.Stats().OpenConnections == 0 {
	// 	ndb, err := mgr.newDB()
	// 	if err == nil {
	// 		db.Close()
	// 		db = ndb
	// 	}
	// }
	select {
	case mgr.pool <- db:
	default:
		mgr.mu.Lock()
		db.Close()
		mgr.activeConn--
		mgr.mu.Unlock()
	}
}

// Get return conn for pool
func (mgr *Manager) Get() (*MySQL, error) {
	switch mgr.opt.mode {
	case AcquireConnModeTimeout:
		return mgr.getConnTimeout()
	case AcquireConnModeBlock:
		return <-mgr.pool, nil
	case AcquireConnModeUnblock:
		return mgr.getConnUnblock()
	default:
		return nil, ErrUnsupportedMode
	}
}

func (mgr *Manager) getConnTimeout() (*MySQL, error) {
	select {
	case conn := <-mgr.pool:
		return conn, nil
	case <-time.After(mgr.opt.waitTimeout):
		return nil, ErrAcquiredConnTimeout
	}
}

func (mgr *Manager) getConnUnblock() (*MySQL, error) {
	select {
	case conn := <-mgr.pool:
		return conn, nil
	default:
		return nil, ErrEmptyConnPool
	}
}

// MySQL conn obj
type MySQL struct {
	db   *sql.DB
	tx   *sql.Tx
	rows *sql.Rows
	rs   sql.Result
	addr *address

	opt *connectionOption
}

func newMySQL(addr *address, opt *connectionOption) *MySQL {
	return &MySQL{
		opt:  opt,
		addr: addr,
	}
}

func (m *MySQL) Info() string {
	return fmt.Sprintf("%v", m.opt)
}

func (m *MySQL) connect() (err error) {
	dsn := fmt.Sprintf("%s:%s@tcp(%s)/%s?charset=%s", m.opt.usr, m.opt.passwd,
		m.addr.String(), m.opt.dbname, m.opt.charset)
	if m.opt.dialTimeout > 0 {
		dsn += fmt.Sprintf("&timeout=%s", m.opt.dialTimeout)
	}
	if m.opt.readTimeout > 0 {
		dsn += fmt.Sprintf("&readTimeout=%s", m.opt.readTimeout)
	}
	if m.opt.writeTimeout > 0 {
		dsn += fmt.Sprintf("&writeTimeout=%s", m.opt.writeTimeout)
	}
	// WTF db_proxy unsupport true
	if m.opt.autoCommit {
		dsn += fmt.Sprintf("&autoCommit=%d", 1)
	}
	if m.opt.loc != "" {
		dsn += fmt.Sprintf("&loc=%s", m.opt.loc)
	}
	if m.opt.parseTime {
		dsn += fmt.Sprintf("&parseTime=%t", m.opt.parseTime)
	}
	if m.opt.columnsWithAlias {
		dsn += fmt.Sprintf("&columnsWithAlias=%t", m.opt.columnsWithAlias)
	}
	m.db, err = sql.Open(mysqlDriver, dsn)
	return
}

// Execute low level api to exec sql
func (m *MySQL) Execute(ctx context.Context, sqlPattern string, args ...interface{}) (err error) {
	et := elapsed.New()
	et.Start()
	if m.opt.debug {
		if m.opt.log == nil {
			fmt.Printf("_mysql||%s||sql:%s values:%v\n", ctx, sqlPattern, args)
		} else {
			m.opt.log.Debugf("_mysql||%s||sql:%s values:%+v", ctx, sqlPattern, args)
		}
	}
	var stmt *sql.Stmt
	if m.rows != nil {
		m.rows.Close()
		m.rows = nil
	}
	if m.tx != nil {
		stmt, err = m.tx.Prepare(sqlPattern)
	} else {
		stmt, err = m.db.Prepare(sqlPattern)
	}
	defer func() {
		if stmt != nil {
			stmt.Close()
		}
		dctx.AddMysqlElapsed(ctx, et.Stop())
	}()
	if err != nil {
		return
	}

	m.rs, err = stmt.Exec(args...)
	return
}

// Query do query
func (m *MySQL) Query(ctx context.Context, sqlPattern string, args ...interface{}) (err error) {
	et := elapsed.New()
	et.Start()
	if m.opt.debug {
		if m.opt.log == nil {
			fmt.Printf("_mysql||%s||sql:%s values:%v\n", ctx, sqlPattern, args)
		} else {
			m.opt.log.Debugf("_mysql||%s||sql:%s values:%+v", ctx, sqlPattern, args)
		}
	}
	if m.rows != nil {
		m.rows.Close()
		m.rows = nil
	}
	if m.tx != nil {
		m.rows, err = m.tx.Query(sqlPattern, args...)
	} else {
		m.rows, err = m.db.Query(sqlPattern, args...)
	}
	dctx.AddMysqlElapsed(ctx, et.Stop())
	return
}

// Begin tr begin
func (m *MySQL) Begin(ctx context.Context) (err error) {
	m.tx, err = m.db.Begin()
	if err != nil {
		return
	}
	if m.rows != nil {
		err = m.rows.Close()
	}
	return
}

// Commit tx commit
func (m *MySQL) Commit(ctx context.Context) (err error) {
	if m.tx == nil {
		return ErrNilTransaction
	}
	if m.rows != nil {
		err = m.rows.Close()
		if err != nil {
			return
		}
	}
	err = m.tx.Commit()
	m.tx = nil
	return
}

// RollBack tx rollback
func (m *MySQL) RollBack(ctx context.Context) (err error) {
	if m.tx == nil {
		return ErrNilTransaction
	}
	if m.rows != nil {
		err = m.rows.Close()
		if err != nil {
			return
		}
	}
	err = m.tx.Rollback()
	m.tx = nil
	return
}

// Insert do
func (m *MySQL) Insert(ctx context.Context, table string, kvPairs map[string]interface{}) (lastID int64, err error) {
	if table == "" {
		return -1, ErrEmptyTable
	}
	length := len(kvPairs)
	if length == 0 {
		return -1, ErrEmptyValues
	}
	keys := make([]string, length)
	pos := make([]string, length)
	values := make([]interface{}, length)
	i := 0
	for field := range kvPairs {
		keys[i] = field
		pos[i] = "?"
		values[i] = kvPairs[field]
		i++
	}
	keyStr := "`" + strings.Join(keys, "`,`") + "`"
	valStr := strings.Join(pos, ",")
	sqlPattern := fmt.Sprintf("INSERT INTO %s(%s) VALUES(%s)", wrapTable(table), keyStr, valStr)
	err = m.Execute(ctx, sqlPattern, values...)
	if err == nil {
		lastID = m.LastInsertID(ctx)
	}
	return
}

// MultiInsert ...
func (m *MySQL) MultiInsert(ctx context.Context, table string, batchData []map[string]interface{}) (lastID int64, err error) {
	et := elapsed.New()
	et.Start()
	if table == "" {
		return -1, ErrEmptyTable
	}
	if len(batchData) == 0 {
		return -1, ErrEmptyValues
	}
	var (
		keys        []string
		placeHolder []string
	)
	for key := range batchData[0] {
		keys = append(keys, key)
		placeHolder = append(placeHolder, "?")
	}
	placeHolderStr := fmt.Sprintf("(%s)", strings.Join(placeHolder, ","))
	keyStr := "`" + strings.Join(keys, "`,`") + "`"

	for _, each := range batchData {
		if len(each) != len(keys) {
			return -1, errNotSameFieldData
		}
	}

	sqlPattern := bytes.NewBufferString(fmt.Sprintf("INSERT INTO %s(%s) VALUES ", wrapTable(table), keyStr))
	for i := range batchData {
		sqlPattern.WriteString(placeHolderStr)
		if i < len(batchData)-1 {
			sqlPattern.WriteString(",")
		}
	}
	var stmt *sql.Stmt
	if m.tx != nil {
		stmt, err = m.tx.Prepare(sqlPattern.String())
	} else {
		stmt, err = m.db.Prepare(sqlPattern.String())
	}
	defer func() {
		if stmt != nil {
			stmt.Close()
		}
		dctx.AddMysqlElapsed(ctx, et.Stop())
	}()

	if err != nil {
		return
	}
	values := make([]interface{}, len(keys)*len(batchData))
	var index = 0
	for _, each := range batchData {
		for _, key := range keys {
			values[index] = each[key]
			index++
		}
	}
	m.rs, err = stmt.Exec(values...)
	if err == nil {
		lastID = m.LastInsertID(ctx)
	}
	return
}

// Upsert ...
func (m *MySQL) Upsert(ctx context.Context, table string, data map[string]interface{}, updateKeys []string) (lastID int64, err error) {
	et := elapsed.New()
	et.Start()
	if table == "" {
		return -1, ErrEmptyTable
	}
	if len(data) == 0 {
		return -1, ErrEmptyValues
	}
	dataKeys := []string{}
	placeHolder := []string{}
	for key := range data {
		dataKeys = append(dataKeys, key)
		placeHolder = append(placeHolder, "?")
	}
	placeHolderStr := fmt.Sprintf("(%s)", strings.Join(placeHolder, ","))
	keyStr := "`" + strings.Join(dataKeys, "`,`") + "`"
	tableName := wrapTable(table)
	sqlPattern := bytes.NewBufferString(
		fmt.Sprintf("INSERT INTO %s(%s) VALUES %s", tableName, keyStr, placeHolderStr))

	if len(updateKeys) > 0 {
		fmt.Fprintf(sqlPattern, " ON DUPLICATE KEY UPDATE ")
		for idx, k := range updateKeys {
			fmt.Fprintf(sqlPattern, "`%s` = VALUES(`%s`)", k, k)
			if idx < len(updateKeys)-1 {
				fmt.Fprintf(sqlPattern, ",")
			}
		}
	}

	var (
		stmt *sql.Stmt
		sql  = sqlPattern.String()
	)
	if m.opt.debug {
		if m.opt.log == nil {
			fmt.Printf("_mysql||%s||sql:%s\n", ctx, sql)
		} else {
			m.opt.log.Debugf("_mysql||%s||sql:%s", ctx, sql)
		}
	}

	if m.tx != nil {
		stmt, err = m.tx.Prepare(sql)
	} else {
		stmt, err = m.db.Prepare(sql)
	}
	defer func() {
		if stmt != nil {
			stmt.Close()
		}
		dctx.AddMysqlElapsed(ctx, et.Stop())
	}()

	if err != nil {
		return
	}
	values := make([]interface{}, len(dataKeys))
	for i, key := range dataKeys {
		values[i] = data[key]
	}
	m.rs, err = stmt.Exec(values...)
	if err == nil {
		lastID = m.LastInsertID(ctx)
	}
	return
}

// Update ...
func (m *MySQL) Update(ctx context.Context, table string, updator map[string]interface{}, condPattern string,
	condArgs ...interface{}) (affect int64, err error) {
	if table == "" {
		return -1, ErrEmptyTable
	}
	length := len(updator)
	if length == 0 {
		return -1, ErrEmptyValues
	}

	var (
		tabName     string
		updatePairs = make([]string, length)
		vals        = make([]interface{}, length)
		i           int
	)
	tabName = wrapTable(table)
	for field := range updator {
		updatePairs[i] = "`" + field + "`=?"
		vals[i] = updator[field]
		i++
	}
	vals = append(vals, condArgs...)
	sqlPattern := fmt.Sprintf("UPDATE %s SET %s", tabName, strings.Join(updatePairs, ","))
	if condPattern != "" {
		sqlPattern += " " + condPattern
	}
	err = m.Execute(ctx, sqlPattern, vals...)
	if err == nil {
		affect = m.AffectRows(ctx)
	}
	return
}

// Delete ...
func (m *MySQL) Delete(ctx context.Context, table string, condPattern string, condArgs ...interface{}) (affect int64, err error) {
	if table == "" {
		return -1, ErrEmptyTable
	}
	if condPattern == "" {
		return -1, ErrEmptyCondition
	}

	var (
		tabName string
	)
	tabName = wrapTable(table)
	sqlPattern := fmt.Sprintf("DELETE FROM %s %s", tabName, condPattern)
	err = m.Execute(ctx, sqlPattern, condArgs...)
	if err == nil {
		affect = m.AffectRows(ctx)
	}
	return
}

// Select ...
func (m *MySQL) Select(ctx context.Context, table string, fields []string, condPattern string, condArgs ...interface{}) error {
	if table == "" {
		return ErrEmptyTable
	}
	if fields == nil || len(fields) == 0 {
		return ErrEmptyValues
	}
	tabName := wrapTable(table)
	fieldStr := ""
	if len(fields) == 1 && fields[0] == matchAllMask {
		fieldStr = fields[0]
	} else {
		fieldStr = "`" + strings.Join(fields, "`,`") + "`"
	}
	sqlPattern := fmt.Sprintf("SELECT %s FROM %s", fieldStr, tabName)
	if condPattern != "" {
		sqlPattern += " " + condPattern
	}
	return m.Query(ctx, sqlPattern, condArgs...)
}

// FetchRow ...
func (m *MySQL) FetchRow(ctx context.Context) (rows Row, err error) {
	rows, err = m.fetchRow(ctx)
	if err == io.EOF {
		err = nil
	}
	return
}

func (m *MySQL) fetchRow(ctx context.Context) (rows Row, err error) {
	if m.rows == nil {
		return nil, io.EOF
	}
	if m.rows.Next() {
		cols, err := m.rows.Columns()
		if err != nil {
			return nil, err
		}
		scanArgs := make([]interface{}, len(cols))
		values := make([]sql.RawBytes, len(cols))
		for i := range values {
			scanArgs[i] = &values[i]
		}
		err = m.rows.Scan(scanArgs...)
		if err != nil {
			return nil, err
		}
		for _, row := range values {
			rows = append(rows, string(row))
		}
		return rows, err
	}
	m.rows.Close()
	return nil, io.EOF
}

// FetchOneRow ...
func (m *MySQL) FetchOneRow(ctx context.Context) (row Row, err error) {
	row, err = m.fetchRow(ctx)
	if m.rows != nil {
		m.rows.Close()
		m.rows = nil
	}
	return
}

// FetchOne ...
func (m *MySQL) FetchOne(ctx context.Context) (firstCol string, err error) {
	row, err := m.FetchOneRow(ctx)
	if err == nil && len(row) > 0 {
		firstCol = row[0]
	}
	return
}

// FetchAll ...
func (m *MySQL) FetchAll(ctx context.Context) (all []Row, err error) {
	for {
		row, err := m.fetchRow(ctx)
		if err != nil && err != io.EOF {
			return nil, err
		}
		if err == io.EOF {
			return all, nil
		}
		all = append(all, row)
	}
}

// FetchRowMap ...
func (m *MySQL) FetchRowMap(ctx context.Context) (rowMap RowMap, err error) {
	rowMap, err = m.fetchRowMap(ctx)
	if err == io.EOF {
		err = nil
	}
	return
}

func (m *MySQL) fetchRowMap(ctx context.Context) (rowMap RowMap, err error) {
	rowMap = make(map[string]string)
	if m.rows == nil {
		return nil, io.EOF
	}
	if m.rows.Next() {
		cols, err := m.rows.Columns()
		if err != nil {
			return nil, err
		}
		scanArgs := make([]interface{}, len(cols))
		values := make([]sql.RawBytes, len(cols))
		for i := range values {
			scanArgs[i] = &values[i]
		}
		err = m.rows.Scan(scanArgs...)
		if err != nil {
			return nil, err
		}
		for i := 0; i < len(cols); i++ {
			rowMap[cols[i]] = string(values[i])
		}
		return rowMap, nil
	}
	m.rows = nil
	return nil, io.EOF
}

// FetchRowMapInterface ...
func (m *MySQL) FetchRowMapInterface(ctx context.Context) (rowMapIntf map[string]interface{}, err error) {
	rowMapIntf, err = m.fetchRowMapInterface(ctx)
	if err == io.EOF {
		err = nil
	}
	return
}

func (m *MySQL) fetchRowMapInterface(ctx context.Context) (rowMapIntf map[string]interface{}, err error) {
	rowMapIntf = make(map[string]interface{})
	if m.rows == nil {
		return nil, io.EOF
	}
	if m.rows.Next() {
		cols, err := m.rows.Columns()
		if err != nil {
			return nil, err
		}
		scanArgs := make([]interface{}, len(cols))
		values := make([]sql.RawBytes, len(cols))
		for i := range values {
			scanArgs[i] = &values[i]
		}
		err = m.rows.Scan(scanArgs...)
		if err != nil {
			return nil, err
		}
		for i := 0; i < len(cols); i++ {
			rowMapIntf[cols[i]] = values[i]
		}
		return rowMapIntf, nil
	}
	m.rows = nil
	return nil, io.EOF
}

// FetchAllMap ...
func (m *MySQL) FetchAllMap(ctx context.Context) (allMap []RowMap, err error) {
	for {
		rowMap, err := m.fetchRowMap(ctx)
		if err != nil && err != io.EOF {
			return nil, err
		}
		if err == io.EOF {
			return allMap, nil
		}
		allMap = append(allMap, rowMap)
	}
}

// LastInsertID ...
func (m *MySQL) LastInsertID(ctx context.Context) (id int64) {
	if m.rs == nil {
		return -1
	}
	id, err := m.rs.LastInsertId()
	if err != nil {
		return -1
	}
	return id
}

// AffectRows ...
func (m *MySQL) AffectRows(ctx context.Context) int64 {
	if m.rs == nil {
		return -1
	}
	affect, err := m.rs.RowsAffected()
	if err != nil {
		return -1
	}
	return affect
}

// Ping ...
func (m *MySQL) Ping() error {
	if m.db == nil {
		return ErrNotOpened
	}
	return m.db.Ping()
}

// Close ...
func (m *MySQL) Close() error {
	if m.rows != nil {
		m.rows.Close()
	}
	m.tx = nil
	m.rs = nil
	m.rows = nil
	return m.db.Close()
}

func wrapTable(table string) string {
	return "`" + strings.Replace(table, ".", "`.`", -1) + "`"
}

// RowMap row result
type RowMap map[string]string

// Row just val
type Row []string

type address struct {
	host string
	port string
}

func newAddress(str string) (*address, error) {
	h, p, err := net.SplitHostPort(str)
	if err != nil {
		return nil, err
	}
	return &address{
		host: h,
		port: p,
	}, nil
}

func (ad *address) String() string {
	return fmt.Sprintf("%s:%s", ad.host, ad.port)
}
