package tiga

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"google.golang.org/protobuf/types/known/timestamppb"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"gorm.io/gorm/schema"
)

type TimestamppbSerializer struct {
}
type JSONField struct {
}
type MySQLDao struct {
	db *gorm.DB
}

type Pagination struct {
	Page     int32
	PageSize int32
	Query    interface{}
	Args     []interface{}
}

func NewMySQLMockDao() (*MySQLDao, sqlmock.Sqlmock) {
	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
	if err != nil {
		panic(err)
	}
	mock.ExpectQuery("SELECT VERSION()").WillReturnRows(sqlmock.NewRows([]string{"version"}).AddRow("5.7.31"))

	dialector := mysql.New(mysql.Config{
		Conn:       db,
		DriverName: "mysql",
	})
	_DB, err := gorm.Open(dialector, &gorm.Config{})
	if err != nil {
		panic(err)
	}
	return &MySQLDao{
		db: _DB,
	}, mock

}
func NewMySQLDao(config *Configuration) *MySQLDao {
	host := config.GetString("mysql.host")
	port := config.GetInt("mysql.port")
	user := config.GetString("mysql.user")
	password := config.GetString("mysql.password")
	database := config.GetString("mysql.database")
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local", user, password, host, port, database)
	// newLogger := logger.New(
	// 	log.New(os.Stdout, "\r\n", log.LstdFlags), // io writer
	// 	logger.Config{
	// 		SlowThreshold:             time.Second,   // Slow SQL threshold
	// 		LogLevel:                  logger.Silent, // Log level
	// 		IgnoreRecordNotFoundError: true,          // Ignore ErrRecordNotFound error for logger
	// 		ParameterizedQueries:      true,          // Don't include params in the SQL log
	// 		Colorful:                  false,         // Disable color
	// 	},
	// )
	db, err := gorm.Open(mysql.New(mysql.Config{
		DSN:                       dsn,   // DSN
		DefaultStringSize:         256,   // string 类型字段的默认长度
		DisableDatetimePrecision:  true,  // 禁用 datetime 精度，MySQL 5.6 之前的数据库不支持
		DontSupportRenameIndex:    true,  // 重命名索引时采用删除并新建的方式，MySQL 5.7 之前的数据库和 MariaDB 不支持重命名索引
		DontSupportRenameColumn:   true,  // 用 `change` 重命名列，MySQL 8 之前的数据库和 MariaDB 不支持重命名列
		SkipInitializeWithVersion: false, // 根据当前 MySQL 版本自动配置

	}), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
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
func (m MySQLDao) RegisterTimeSerializer() {
	schema.RegisterSerializer("timepb", TimestamppbSerializer{})
	schema.RegisterSerializer("json", JSONField{})

}
func (m MySQLDao) Save(model interface{}) error {
	return m.db.Save(model).Error
}
func (m MySQLDao) Create(ctx context.Context, model interface{}) error {
	return m.db.WithContext(ctx).Create(model).Error
}
func (m MySQLDao) Update(ctx context.Context, model interface{}, value interface{}) error {
	return m.db.WithContext(ctx).Model(model).Updates(value).Error
}
func (m MySQLDao) UpdateColumns(model interface{}, value interface{}) error {
	return m.db.Where(model).Updates(value).Error
}
func (m MySQLDao) UpdateWithQuery(model interface{}, value interface{}, fields []string, query interface{}, args ...interface{}) error {
	return m.db.Model(model).Where(query, args...).Select(fields).Updates(value).Error
}
func (m MySQLDao) UpdateSelectColumns(ctx context.Context, where interface{}, value interface{}, selectCol ...string) error {
	return m.db.WithContext(ctx).Where(where).Select(selectCol).Updates(value).Error
}
func (m MySQLDao) Pagination(ctx context.Context, model interface{}, pagination *Pagination) error {
	return m.db.WithContext(ctx).Where(pagination.Query, pagination.Args...).Limit(int(pagination.PageSize)).Offset(int(pagination.PageSize * (pagination.Page - 1))).Find(model).Error
}
func (m MySQLDao) Delete(model interface{}, conds ...interface{}) error {
	return m.db.Delete(model, conds...).Error
}
func (m MySQLDao) BatchUpdates(models []interface{}, selectCol ...string) error {
	tx := m.db.Begin()
	if tx.Error != nil {
		return tx.Error
	}
	for _, model := range models {
		tx.Select(selectCol).Updates(model)
	}
	return tx.Commit().Error

}
func (m MySQLDao) Find(ctx context.Context, model interface{}, query interface{}, args ...interface{}) error {
	return m.db.WithContext(ctx).Where(query, args...).Find(model).Error
}
func (m MySQLDao) First(model interface{}, query interface{}, args ...interface{}) error {
	return m.db.Where(query, args...).First(model).Error
}
func (m MySQLDao) FirstWithMuiltQuery(model interface{}, query *gorm.DB) error {
	return query.First(model).Error
}
func (m MySQLDao) Last(model interface{}, query interface{}, args ...interface{}) error {
	return m.db.Where(query, args...).Last(model).Error
}
func (m MySQLDao) FindAll(model interface{}, query interface{}, args ...interface{}) error {
	return m.db.Where(query, args...).Find(model).Error
}
func (m MySQLDao) AutoMigrate(model interface{}) error {
	return m.db.AutoMigrate(model)
}
func (m MySQLDao) Count(model interface{}, query interface{}, args ...interface{}) (int64, error) {
	var count int64
	err := m.db.Where(query, args...).Model(model).Count(&count).Error
	return count, err
}
func (m MySQLDao) CreateInBatches(models interface{}) error {
	size, err := GetElementCount(models)
	if err != nil {
		return err
	}
	return m.db.CreateInBatches(models, size).Error
}
func (m MySQLDao) Where(query interface{}, args ...interface{}) *gorm.DB {
	return m.db.Where(query, args...)
}
func (m MySQLDao) GroupAndCount(model interface{}, result interface{}, groupField string, selectQuery string, query interface{}, args ...interface{}) error {
	err := m.db.Model(model).Select(selectQuery).Where(query, args...).Group(groupField).Find(result).Error
	return err
}
func (m MySQLDao) Begin(opts ...*sql.TxOptions) *gorm.DB {
	return m.db.Begin(opts...)
}
func (m MySQLDao) GetModel(model interface{}) *gorm.DB {
	return m.db.Model(model)
}
func (m MySQLDao) GetTable(name string, args ...interface{}) *gorm.DB {
	return m.db.Table(name, args...)
}
func (m MySQLDao) TableName(model interface{}) (string, error) {
	err := m.db.Statement.Parse(model)
	if err != nil {
		return "", err
	}
	return m.db.Statement.Table, nil
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
	field.ReflectValueOf(ctx, dst).Set(reflect.ValueOf(timestamp))
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
	dbTime := timestamp.AsTime().UTC()

	return dbTime, nil
}

func (s JSONField) Scan(ctx context.Context, field *schema.Field, dst reflect.Value, dbValue interface{}) error {
	// 确认 dbValue 是 time.Time 类型
	if dbValue == nil {
		return nil // 如果 dbValue 是 nil，没有什么要设置的
	}

	// 断言 dbValue 的类型是 datatypes.JSON
	dbJSON, ok := dbValue.([]byte)
	if !ok {
		return errors.New("dbJSON is not a datatypes.JSON type")
	}

	// 将 time.Time 转换为 *timestamppb.Timestamp
	var stringArray []string
	if err := json.Unmarshal(dbJSON, &stringArray); err != nil {

		return err
	}
	field.ReflectValueOf(ctx, dst).Set(reflect.ValueOf(stringArray))
	return nil
}

func (s JSONField) Value(ctx context.Context, field *schema.Field, dst reflect.Value, fieldValue interface{}) (interface{}, error) {
	// 确认 value 是 *timestamppb.Timestamp 类型
	if fieldValue == nil {
		return nil, nil // 如果 value 是 nil，没有什么要设置的
	}
	return InterfaceToBytes(fieldValue)
}
