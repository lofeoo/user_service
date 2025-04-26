package db

import (
	"fmt"
	"time"

	"github.com/cloudwego/hertz/pkg/common/hlog"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"gorm.io/gorm/schema"

	"user_service/internal/pkg/conf"
)

var (
	DB *gorm.DB
)

// 初始化MySQL连接
func InitMySQL(config *conf.MySQLConfig) (*gorm.DB, error) {
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		config.User,
		config.Password,
		config.Host,
		config.Port,
		config.Database,
	)

	// 配置日志
	gormLogger := logger.New(
		&mysqlLogger{},
		logger.Config{
			SlowThreshold: time.Second, // 慢查询阈值
			LogLevel:      logger.Info,
			Colorful:      false,
		},
	)

	// 连接配置
	gormConfig := &gorm.Config{
		Logger: gormLogger,
		NamingStrategy: schema.NamingStrategy{
			SingularTable: true, // 使用单数表名
		},
		DisableForeignKeyConstraintWhenMigrating: true, // 禁用外键
	}

	// 建立连接
	db, err := gorm.Open(mysql.Open(dsn), gormConfig)
	if err != nil {
		hlog.Errorf("MySQL连接失败: %v", err)
		return nil, err
	}

	// 设置连接池
	sqlDB, err := db.DB()
	if err != nil {
		hlog.Errorf("获取底层数据库连接失败: %v", err)
		return nil, err
	}

	// 设置连接池参数
	sqlDB.SetMaxIdleConns(config.MaxIdleConn)
	sqlDB.SetMaxOpenConns(config.MaxOpenConn)
	sqlDB.SetConnMaxLifetime(time.Hour)

	// 测试连接
	if err := sqlDB.Ping(); err != nil {
		hlog.Errorf("MySQL测试连接失败: %v", err)
		return nil, err
	}

	hlog.Infof("MySQL连接成功: %s:%d/%s", config.Host, config.Port, config.Database)
	DB = db
	return db, nil
}

// 自定义MySQL日志
type mysqlLogger struct{}

func (l *mysqlLogger) Printf(format string, args ...interface{}) {
	hlog.Debugf(format, args...)
}
