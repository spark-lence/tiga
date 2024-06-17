package tiga

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"strings"
	"sync"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"google.golang.org/protobuf/types/known/timestamppb"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"gorm.io/gorm/logger"
	"gorm.io/gorm/schema"
)

type TimestamppbSerializer struct {
}
type JSONField struct {
}
type MySQLDao struct {
	db  *gorm.DB
	cfg *Configuration
}

type Pagination struct {
	Page     int32
	PageSize int32
	Query    interface{}
	Args     []interface{}
	Select   []string
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
	prefix := config.GetString("mysql.table_prefix")
	password := config.GetString("mysql.password")
	database := config.GetString("mysql.database")
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local", user, password, host, port, database)

	db, err := gorm.Open(mysql.New(mysql.Config{
		DSN:                       dsn,   // DSN
		DefaultStringSize:         256,   // string 类型字段的默认长度
		DisableDatetimePrecision:  true,  // 禁用 datetime 精度，MySQL 5.6 之前的数据库不支持
		DontSupportRenameIndex:    true,  // 重命名索引时采用删除并新建的方式，MySQL 5.7 之前的数据库和 MariaDB 不支持重命名索引
		DontSupportRenameColumn:   true,  // 用 `change` 重命名列，MySQL 8 之前的数据库和 MariaDB 不支持重命名列
		SkipInitializeWithVersion: false, // 根据当前 MySQL 版本自动配置

	}), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
		NamingStrategy: schema.NamingStrategy{
			TablePrefix: prefix, // 表名前缀，`User` 的表名应该是 `tiga_users`
		},
	})
	if err != nil {
		panic(err)
	}
	return &MySQLDao{
		db:  db,
		cfg: config,
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
func (m MySQLDao) Upsert(ctx context.Context, model interface{}, updateSelect ...string) error {

	prefix := m.cfg.GetString("mysql.table_prefix")

	s, err := schema.Parse(model, &sync.Map{}, schema.NamingStrategy{TablePrefix: prefix})
	if err != nil {
		return err

	}
	structNames := make(map[string]string)
	for _, field := range s.Fields {
		structNames[field.StructField.Name] = field.DBName
	}
	modelType := reflect.TypeOf(model).Elem()
	fields := make([]string, 0)
	if len(updateSelect) > 0 {
		fields = append(fields, updateSelect...)
	} else {
		for i := 0; i < modelType.NumField(); i++ {
			field := modelType.Field(i).Name
			if field != "CreatedAt" && modelType.Field(i).Tag.Get("gorm") != "" && modelType.Field(i).Tag.Get("gorm") != "-" { // 确保排除CreatedAt字段
				fields = append(fields, structNames[field])
			}
		}
	}
	return m.db.Clauses(clause.OnConflict{
		DoUpdates: clause.AssignmentColumns(fields),
	}).Create(model).Error
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
	base := m.db.WithContext(ctx)
	if len(pagination.Select) > 0 {
		base = base.Select(pagination.Select)
	}
	return base.Where(pagination.Query, pagination.Args...).Limit(int(pagination.PageSize)).Offset(int(pagination.PageSize * (pagination.Page - 1))).Find(model).Error
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
func (m MySQLDao) First(ctx context.Context, model interface{}, query interface{}, args ...interface{}) error {
	return m.db.WithContext(ctx).Where(query, args...).First(model).Error
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
func (m MySQLDao) getRenames(model interface{}) (map[string]string, error) {
	// 获取结构体类型
	modelType := reflect.TypeOf(model)
	// modelVal := reflect.ValueOf(model)
	if modelType.Kind() == reflect.Ptr {
		modelType = modelType.Elem()
		// modelVal = modelVal.Elem()

	}
	if modelType.Kind() != reflect.Struct {
		return nil, fmt.Errorf("%s not a struct type", modelType.Kind().String())
	}
	names := make(map[string]string)

	// 遍历结构体的字段
	for i := 0; i < modelType.NumField(); i++ {
		field := modelType.Field(i)
		// 检查字段名称是否匹配
		if oldName := field.Tag.Get("old_column"); oldName != "" {
			// 获取字段值
			tag := field.Tag.Get("gorm")
			// 	// 解析GORM标签
			tagParts := strings.Split(tag, ";")
			for _, part := range tagParts {
				kv := strings.Split(part, ":")
				if len(kv) == 2 && strings.TrimSpace(kv[0]) == "column" {
					names[oldName] = strings.TrimSpace(kv[1])
					// value := modelVal.Field(i).Interface()
					// return strings.TrimSpace(kv[1]), value, nil
				}
			}
		}

	}
	return names, nil
	// return nil, fmt.Errorf("not found primary column")
}
func (m MySQLDao) AutoMigrate(model interface{}) error {
	// m.db.Migrator().r
	names, err := m.getRenames(model)
	if err != nil {
		return err
	}
	if m.db.Migrator().HasTable(model) {
		for oldName, newName := range names {
			if err = m.db.Migrator().RenameColumn(model, oldName, newName); err != nil && !strings.Contains(err.Error(), "Unknown column") {
				return fmt.Errorf("rename column %s to %s failed:%w", oldName, newName, err)
			}
		}
	}
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
