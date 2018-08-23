package dbutil

import (
	"errors"
	"fmt"
	"github.com/astaxie/beego/orm"
	_ "github.com/go-sql-driver/mysql"
	"github.com/qjpcpu/log/logging"
	"reflect"
	"runtime/debug"
)

type ConnOption struct {
	Conn        string
	Alias       string // default is "default"
	MaxIdleConn int
	MaxOpenConn int
}

func InitMysql(options ...ConnOption) error {
	if len(options) == 0 {
		return errors.New("no db connection string")
	}
	orm.RegisterDriver("mysql", orm.DRMySQL)
	for _, opt := range options {
		if opt.Alias == "" {
			opt.Alias = "default"
		}
		var params []int
		if opt.MaxIdleConn > 0 || opt.MaxOpenConn > 0 {
			params = append(params, opt.MaxIdleConn)
			if opt.MaxOpenConn > 0 {
				params = append(params, opt.MaxOpenConn)
			}
		}
		if err := orm.RegisterDataBase(opt.Alias, "mysql", opt.Conn, params...); err != nil {
			return err
		}
	}
	return nil
}

func SetDBLog(file_path string) error {
	flog, err := logging.NewFileLogWriter(file_path, logging.RotateDaily)
	if err != nil {
		return fmt.Errorf("set db log fail:%v", err)
	}
	orm.Debug = true
	orm.DebugLog = orm.NewLog(flog)
	return nil
}

func SetDefaultRowsLimit(limit int) {
	orm.DefaultRowsLimit = limit
}

func RegisterModel(models ...interface{}) {
	orm.RegisterModel(models...)
}

func GetOrm() Ormer {
	o := orm.NewOrm()
	return o
}

func GetOrmOf(alias string) Ormer {
	o := orm.NewOrm()
	o.Using(alias)
	return o
}

type Task func() error

func errHandler(task Task) (err error) {
	defer func() {
		if e := recover(); e != nil {
			orm.DebugLog.Printf("panic: %v; calltrace:%s", e, string(debug.Stack()))
			err = fmt.Errorf("%v", e)
		}
	}()
	return task()
}

func ExecTransaction(o Ormer, transction ...Task) error {
	if err := o.Begin(); err != nil {
		orm.DebugLog.Printf("DB begin transaction failed: %s", err.Error())
		return err
	}

	for _, task := range transction {
		if task != nil {
			if err := errHandler(task); err != nil {
				if rberr := o.Rollback(); rberr != nil {
					orm.DebugLog.Printf("DB rollback transaction failed: %s", rberr.Error())
				}
				return err
			}
		}
	}

	if err := o.Commit(); err != nil {
		o.Rollback()
		orm.DebugLog.Printf("DB commit transaction failed: %s", err.Error())
		return err
	}
	return nil
}

// convenient alias
type Ormer = orm.Ormer
type Params = orm.Params

const (
	ColAdd      = orm.ColAdd
	ColMinus    = orm.ColMinus
	ColMultiply = orm.ColMultiply
	ColExcept   = orm.ColExcept
)

// function/variables reexport
var (
	ColValue  = orm.ColValue
	ErrNoRows = orm.ErrNoRows
)

// helpers
func IsNoRowsErr(err error) bool {
	return err != nil && err == orm.ErrNoRows
}

// ToArgsSlice([]string{"x","y","z"}) ==> []inteface{}{"x","y","z"}
func ToArgsSlice(array interface{}) []interface{} {
	v := reflect.ValueOf(array)
	if v.Kind() != reflect.Slice && v.Kind() != reflect.Array {
		panic("input should be array/slice")
	}
	size := v.Len()
	resarray := make([]interface{}, size)
	for i := 0; i < size; i++ {
		resarray[i] = v.Index(i).Interface()
	}
	return resarray
}

// Placeholders([]string{"a","b","c"}) ==> ?,?,?
func Placeholders(array interface{}) string {
	v := reflect.ValueOf(array)
	if v.Kind() != reflect.Slice && v.Kind() != reflect.Array {
		panic("input should be array/slice")
	}
	size := v.Len()
	if size == 0 {
		panic("empty array")
	}
	b_question := byte(63) // ?
	b_comma := byte(44)    //,
	holders := make([]byte, size+size-1)
	var j int
	for i := 0; i < size; i++ {
		holders[j] = b_question
		j++
		if i == size-1 {
			break
		}
		holders[j] = b_comma
		j++
	}
	return string(holders)
}
