package tiga

import (
	"context"
	"errors"
	"reflect"
	"time"

	"google.golang.org/protobuf/types/known/timestamppb"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/schema"
)
type TimestamppbSerializer struct {
}
type MySQLDao struct {
	db *gorm.DB
}

func NewMySQLDao(dsn string) *MySQLDao {
	db, err := gorm.Open(mysql.New(mysql.Config{
		DSN:                       dsn,   // DSN
		DefaultStringSize:         256,   // string 类型字段的默认长度
		DisableDatetimePrecision:  true,  // 禁用 datetime 精度，MySQL 5.6 之前的数据库不支持
		DontSupportRenameIndex:    true,  // 重命名索引时采用删除并新建的方式，MySQL 5.7 之前的数据库和 MariaDB 不支持重命名索引
		DontSupportRenameColumn:   true,  // 用 `change` 重命名列，MySQL 8 之前的数据库和 MariaDB 不支持重命名列
		SkipInitializeWithVersion: false, // 根据当前 MySQL 版本自动配置
	}), &gorm.Config{})
	if err != nil {
		panic(err)
	}
	return &MySQLDao{
		db: db,
	}
}
func (m MySQLDao) Close() error {
	db, err := m.db.DB()
	if err != nil {
		return err
	}
	return db.Close()
}
func (m MySQLDao) RegisterTtimepbSerializer() {
	schema.RegisterSerializer("timepb", TimestamppbSerializer{})

}
func (m MySQLDao) Save(model interface{}) error {
	return m.db.Save(model).Error
}
func (m MySQLDao) Create(model interface{}) error {
	return m.db.Create(model).Error
}
func (m MySQLDao) Update(model interface{}) error {
	return m.db.Updates(model).Error
}
func (m MySQLDao) Delete(model interface{}) error {
	return m.db.Delete(model).Error
}
func (m MySQLDao) Find(model interface{}, query interface{}, args ...interface{}) error {
	return m.db.Where(query, args...).Find(model).Error
}
func (m MySQLDao) First(model interface{}, query interface{}, args ...interface{}) error {
	return m.db.Where(query, args...).First(model).Error
}
func (m MySQLDao) Last(model interface{}, query interface{}, args ...interface{}) error {
	return m.db.Where(query, args...).Last(model).Error
}
func (m MySQLDao) Count(model interface{}, query interface{}, args ...interface{}) (int64, error) {
	var count int64
	err := m.db.Where(query, args...).Model(model).Count(&count).Error
	return count, err
}



func (s TimestamppbSerializer) Scan(ctx context.Context, field *schema.Field, dst reflect.Value, dbValue interface{}) error {
	// 确认 dbValue 是 time.Time 类型
	if dbValue == nil {
		return nil // 如果 dbValue 是 nil，没有什么要设置的
	}

	// 断言 dbValue 的类型是 time.Time
	dbTime, ok := dbValue.(time.Time)
	if !ok {
		return errors.New("dbValue is not a time.Time type")
	}

	// 将 time.Time 转换为 *timestamppb.Timestamp
	timestamp := timestamppb.New(dbTime)

	// 确保 dst 可以设置值
	if dst.CanSet() {
		// 设置转换后的值
		dst.Set(reflect.ValueOf(timestamp))
	} else {
		return errors.New("destination cannot be set")
	}
	return nil
}

func (s TimestamppbSerializer) Value(ctx context.Context, field *schema.Field, dst reflect.Value, fieldValue interface{}) (interface{}, error) {
	// 确认 value 是 *timestamppb.Timestamp 类型
	if fieldValue == nil {
		return nil, nil // 如果 value 是 nil，没有什么要设置的
	}

	// 断言 value 的类型是 *timestamppb.Timestamp
	timestamp, ok := fieldValue.(*timestamppb.Timestamp)
	if !ok {
		return nil, errors.New("value is not a *timestamppb.Timestamp type")
	}

	// 将 *timestamppb.Timestamp 转换为 time.Time
	dbTime := timestamp.AsTime()

	return dbTime, nil
}
