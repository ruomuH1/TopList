package Common

import (
	"github.com/tophubs/TopList/Config"
	"database/sql"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"log"
	"math"
	"strconv"
	"strings"
	"sync"
	"time"
)

var DbPool *sync.Pool
var GlobalDb *sql.DB

type MySql struct {
	source      string  // 数据库源
	driver      string  // 数据库驱动
	fields      string  // 字段
	tableName   string  // 表名
	whereStr    string  // where语句
	limitNumber string  // 限制条数
	orderBy     string  // 排序条件
	execStr     string  // 执行sql语句
	conn        *sql.DB // 数据库连接
}

type MysqlCfg struct {
	Source, Driver string
}

// 初始化连接池
func init() {
	MySql := MySql{}
	MySql.source = Config.MySql().Source
	MySql.driver = Config.MySql().Driver
	
	// 先连接到 MySQL 服务器（不指定数据库）
	tempSource := strings.Replace(MySql.source, "/mine", "/", 1)
	db, err := sql.Open(MySql.driver, tempSource)
	MySql.checkErr(err)
	
	// 创建数据库
	_, err = db.Exec("CREATE DATABASE IF NOT EXISTS mine CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci")
	MySql.checkErr(err)
	
	// 关闭临时连接
	db.Close()
	
	// 重新连接到 mine 数据库
	db, err = sql.Open(MySql.driver, MySql.source)
	db.SetMaxOpenConns(2000)             // 最大链接
	db.SetMaxIdleConns(1000)             // 空闲连接，也就是连接池里面的数量
	db.SetConnMaxLifetime(7 * time.Hour) // 设置最大生成周期是7个小时
	MySql.checkErr(err)
	
	// 创建表结构
	createTableSQL := `
	CREATE TABLE IF NOT EXISTS hotData2 (
	  id int(11) NOT NULL AUTO_INCREMENT,
	  str text CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci,
	  dataType varchar(45) CHARACTER SET utf8mb4 COLLATE utf8mb4_bin NOT NULL,
	  name varchar(45) CHARACTER SET utf8mb4 COLLATE utf8mb4_bin NOT NULL,
	  isShow int(11) DEFAULT '0',
	  rss text,
	  PRIMARY KEY (id),
	  KEY hotData2__index_key (dataType),
	  KEY hotData2__index_name (name)
	) ENGINE=InnoDB AUTO_INCREMENT=114 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
	`
	_, err = db.Exec(createTableSQL)
	MySql.checkErr(err)
	
	// 插入初始数据
	insertDataSQL := `
	INSERT IGNORE INTO hotData2 (id, str, dataType, name, isShow, rss) VALUES
	(1, '{"Code":0,"Message":"获取成功","Data":[{"title":"如何看待环球时报记者付国豪在香港机场的遭遇及事件后续?","url":"https://www.zhihu.com/question/340424050"},{"title":"如何看待在最新的 iOS13 测试版中，香港（中国）被显示为香港？","url":"https://www.zhihu.com/question/340106103"}]}', 'ZhiHu', '知乎', 1, NULL),
	(2, '{"Code":0,"Message":"获取成功","Data":[]}', 'V2EX', 'V2EX', 1, NULL),
	(3, '{"Code":0,"Message":"获取成功","Data":[]}', 'WeiBo', '微博', 1, NULL),
	(4, '{"Code":0,"Message":"获取成功","Data":[]}', 'TieBa', '贴吧', 1, NULL),
	(5, '{"Code":0,"Message":"获取成功","Data":[]}', 'DouBan', '豆瓣', 1, NULL),
	(6, '{"Code":0,"Message":"获取成功","Data":[]}', 'TianYa', '天涯', 1, NULL),
	(7, '{"Code":0,"Message":"获取成功","Data":[]}', 'HuPu', '虎扑', 1, NULL),
	(8, '{"Code":0,"Message":"获取成功","Data":[]}', 'GitHub', 'GitHub', 1, NULL),
	(9, '{"Code":0,"Message":"获取成功","Data":[]}', 'BaiDu', '百度', 1, NULL),
	(10, '{"Code":0,"Message":"获取成功","Data":[]}', '36Kr', '36氪', 1, NULL),
	(11, '{"Code":0,"Message":"获取成功","Data":[]}', 'QDaily', '好奇心日报', 1, NULL),
	(12, '{"Code":0,"Message":"获取成功","Data":[]}', 'GuoKr', '果壳', 1, NULL),
	(13, '{"Code":0,"Message":"获取成功","Data":[]}', 'HuXiu', '虎嗅', 1, NULL),
	(14, '{"Code":0,"Message":"获取成功","Data":[]}', 'ZHDaily', '知乎日报', 1, NULL),
	(15, '{"Code":0,"Message":"获取成功","Data":[]}', 'Segmentfault', 'SegmentFault', 1, NULL),
	(16, '{"Code":0,"Message":"获取成功","Data":[]}', 'WYNews', '网易新闻', 1, NULL),
	(17, '{"Code":0,"Message":"获取成功","Data":[]}', 'WaterAndWood', '水木社区', 1, NULL),
	(18, '{"Code":0,"Message":"获取成功","Data":[]}', 'HacPai', '黑客派', 1, NULL),
	(19, '{"Code":0,"Message":"获取成功","Data":[]}', 'KD', '凯迪社区', 1, NULL),
	(20, '{"Code":0,"Message":"获取成功","Data":[]}', 'NGA', 'NGA', 1, NULL),
	(21, '{"Code":0,"Message":"获取成功","Data":[]}', 'WeiXin', '微信', 1, NULL),
	(22, '{"Code":0,"Message":"获取成功","Data":[]}', 'Mop', '猫扑', 1, NULL),
	(23, '{"Code":0,"Message":"获取成功","Data":[]}', 'Chiphell', 'Chiphell', 1, NULL),
	(24, '{"Code":0,"Message":"获取成功","Data":[]}', 'JianDan', '煎蛋', 1, NULL),
	(25, '{"Code":0,"Message":"获取成功","Data":[]}', 'ChouTi', '抽屉', 1, NULL),
	(26, '{"Code":0,"Message":"获取成功","Data":[]}', 'ITHome', 'IT之家', 1, NULL);
	`
	_, err = db.Exec(insertDataSQL)
	MySql.checkErr(err)
	
	GlobalDb = db
}

/**
sql.Open函数实际上是返回一个连接池对象，不是单个连接。
在open的时候并没有去连接数据库，只有在执行query、exce方法的时候才会去实际连接数据库。
在一个应用中同样的库连接只需要保存一个sql.Open之后的db对象就可以了，不需要多次open。
*/
//func CreateConn() interface{} {
//	MySql := MySql{}
//	var cfg Config.Config
//	cfg = new(Config.Mysql)
//	MySql.source = cfg.GetConfig()["source"].(string)
//	MySql.driver = cfg.GetConfig()["driver"].(string)
//	db, err := sql.Open(MySql.driver, MySql.source)
//	db.SetMaxOpenConns(2000)  // 最大链接
//	db.SetMaxIdleConns(1000)  // 空间连接，也就是连接池里面的数量
//	MySql.checkErr(err)
//	MySql.conn = db
//	return db
//}

func (MySql MySql) GetConn() *MySql {
	MySql.conn = GlobalDb
	return &MySql
}

func (MySql *MySql) Close() error {
	err := MySql.conn.Close()
	if err != nil {
		return err
	}
	return nil
}

/**
查询方法
*/
func (MySql *MySql) Select(tableName string, field []string) *MySql {
	var allField string
	allField = strings.Join(field, ",")
	MySql.fields = "select " + allField + " from " + tableName
	MySql.tableName = tableName
	return MySql
}

/**
where子句
*/
func (MySql *MySql) Where(cond map[string]string) *MySql {
	var whereStr = ""
	if len(cond) != 0 {
		whereStr = " where "
		for key, value := range cond {
			if !strings.Contains(key, "=") && !strings.Contains(key, ">") && !strings.Contains(key, "<") {
				key += "="
			}
			whereStr += key + "'" + value + "'" + " AND "
		}
	}
	// 删除所有字段最后一个,
	whereStr = strings.TrimSuffix(whereStr, "AND ")
	MySql.whereStr = whereStr
	return MySql
}

func (MySql *MySql) Limit(number int) *MySql {
	MySql.limitNumber = " limit " + strconv.Itoa(number)
	return MySql
}

func (MySql *MySql) OrderByString(orderString ...string) *MySql {
	if len(orderString) > 2 || len(orderString) <= 0 {
		log.Fatal("传入参数错误")
	} else if len(orderString) == 1 {
		MySql.orderBy = " ORDER BY " + orderString[0] + " ASC"
	} else {
		MySql.orderBy = " ORDER BY " + orderString[0] + " " + orderString[1]
	}
	return MySql
}

/**
更新方法
*/
func (MySql MySql) Update(tableName string, str map[string]string) int64 {
	var tempStr = ""
	var allValue []interface{}
	for key, value := range str {
		tempStr += key + "=" + "?" + ","
		allValue = append(allValue, value)
	}
	tempStr = strings.TrimSuffix(tempStr, ",")
	MySql.execStr = "update " + tableName + " set " + tempStr
	var allStr = MySql.execStr + MySql.whereStr
	stmt, err := MySql.conn.Prepare(allStr)
	MySql.checkErr(err)
	res, err := stmt.Exec(allValue...)
	MySql.checkErr(err)
	rows, err := res.RowsAffected()
	return rows

}

/**
删除方法
*/
func (MySql MySql) Delete(tableName string) int64 {
	var tempStr = ""
	tempStr = "delete from " + tableName + MySql.whereStr
	fmt.Println(tempStr)
	stmt, err := MySql.conn.Prepare(tempStr)
	MySql.checkErr(err)
	res, err := stmt.Exec()
	MySql.checkErr(err)
	rows, err := res.RowsAffected()
	return rows
}

/**
插入方法
*/
func (MySql MySql) Insert(tableName string, data map[string]string) int64 {
	var allField = ""
	var allValue = ""
	var allTrueValue []interface{}
	if len(data) != 0 {
		for key, value := range data {
			allField += key + ","
			allValue += "?" + ","
			allTrueValue = append(allTrueValue, value)
		}
	}
	allValue = strings.TrimSuffix(allValue, ",")
	allField = strings.TrimSuffix(allField, ",")
	allValue = "(" + allValue + ")"
	allField = "(" + allField + ")"
	var theStr = "insert into " + tableName + " " + allField + " values " + allValue
	stmt, err := MySql.conn.Prepare(theStr)
	MySql.checkErr(err)
	res, err := stmt.Exec(allTrueValue...)
	if err != nil {
		fmt.Println(err.Error())
		return 0
	}
	MySql.checkErr(err)
	id, err := res.LastInsertId()
	return id
}

/**
分页查询
*/
func (MySql MySql) Pagination(Page int, Limit int) map[string]interface{} {
	res := MySql.GetConn().Select(MySql.tableName, []string{"count(*) as count"}).QueryRow()
	count, _ := strconv.Atoi(res["count"])
	// 计算总页码数
	totalPage := int(math.Ceil(float64(count) / float64(Limit)))
	if Page > totalPage {
		Page = totalPage
	}
	if Page <= 0 {
		Page = 1
	}
	// 计算偏移量
	setOff := (Page - 1) * Limit
	queryStr := MySql.fields + MySql.whereStr + MySql.orderBy + " limit " + strconv.Itoa(setOff) + "," + strconv.Itoa(Limit)
	rows, err := MySql.conn.Query(queryStr)
	defer rows.Close()
	MySql.checkErr(err)
	Column, err := rows.Columns()
	MySql.checkErr(err)
	// 创建一个查询字段类型的slice
	values := make([]sql.RawBytes, len(Column))
	// 创建一个任意字段类型的slice
	scanArgs := make([]interface{}, len(values))
	// 创建一个slice保存所以的字段
	var allRows []interface{}
	for i := range values {
		// 把values每个参数的地址存入scanArgs
		scanArgs[i] = &values[i]
	}
	for rows.Next() {
		// 把存放字段的元素批量放进去
		err = rows.Scan(scanArgs...)
		MySql.checkErr(err)
		tempRow := make(map[string]string, len(Column))
		for i, col := range values {
			var key = Column[i]
			tempRow[key] = string(col)
		}
		allRows = append(allRows, tempRow)
	}
	returnData := make(map[string]interface{})
	returnData["totalPage"] = totalPage
	returnData["currentPage"] = Page
	returnData["rows"] = allRows
	return returnData
}

func (MySql MySql) QueryAll() []map[string]string {
	var queryStr = MySql.fields + MySql.whereStr + MySql.orderBy + MySql.limitNumber
	rows, err := MySql.conn.Query(queryStr)
	defer rows.Close()
	MySql.checkErr(err)
	Column, err := rows.Columns()
	MySql.checkErr(err)
	// 创建一个查询字段类型的slice
	values := make([]sql.RawBytes, len(Column))
	// 创建一个任意字段类型的slice
	scanArgs := make([]interface{}, len(values))
	// 创建一个slice保存所以的字段
	var allRows []map[string]string
	for i := range values {
		// 把values每个参数的地址存入scanArgs
		scanArgs[i] = &values[i]
	}
	for rows.Next() {
		// 把存放字段的元素批量放进去
		err = rows.Scan(scanArgs...)
		MySql.checkErr(err)
		tempRow := make(map[string]string, len(Column))
		for i, col := range values {
			var key = Column[i]
			tempRow[key] = string(col)
		}
		allRows = append(allRows, tempRow)
	}
	return allRows
}

func (MySql MySql) ExecSql(queryStr string) []map[string]string {
	rows, err := MySql.conn.Query(queryStr)
	defer rows.Close()
	MySql.checkErr(err)
	Column, err := rows.Columns()
	MySql.checkErr(err)
	// 创建一个查询字段类型的slice
	values := make([]sql.RawBytes, len(Column))
	// 创建一个任意字段类型的slice
	scanArgs := make([]interface{}, len(values))
	// 创建一个slice保存所以的字段
	var allRows []map[string]string
	for i := range values {
		// 把values每个参数的地址存入scanArgs
		scanArgs[i] = &values[i]
	}
	for rows.Next() {
		// 把存放字段的元素批量放进去
		err = rows.Scan(scanArgs...)
		MySql.checkErr(err)
		tempRow := make(map[string]string, len(Column))
		for i, col := range values {
			var key = Column[i]
			tempRow[key] = string(col)
		}
		allRows = append(allRows, tempRow)
	}
	return allRows
}

/**
查询单行
*/
func (MySql MySql) QueryRow() map[string]string {
	var queryStr = MySql.fields + MySql.whereStr + MySql.orderBy + MySql.limitNumber
	result, err := MySql.conn.Query(queryStr)
	defer result.Close()
	MySql.checkErr(err)
	Column, err := result.Columns()
	// 创建一个查询字段类型的slice的键值对
	values := make([]sql.RawBytes, len(Column))
	// 创建一个任意字段类型的slice的键值对
	scanArgs := make([]interface{}, len(values))
	for i := range values {
		// 把values每个参数的地址存入scanArgs
		scanArgs[i] = &values[i]
	}

	for result.Next() {
		err = result.Scan(scanArgs...)
		MySql.checkErr(err)
	}
	tempRow := make(map[string]string, len(Column))
	for i, col := range values {
		var key = Column[i]
		tempRow[key] = string(col)
	}
	return tempRow

}

/**
检查错误
*/
func (MySql MySql) checkErr(err error) {
	if err != nil {
		log.Fatal("错误：", err)
	}
}
