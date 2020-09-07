package main

import (
	"database/sql/driver"
	"flag"
	"fmt"
	"net/http"
	"path/filepath"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"

	_ "github.com/jinzhu/gorm/dialects/sqlite"
)

var (
	sqlite *gorm.DB
	dbPath string
)

// Restful 简单的 restful api 返回结构
type Restful struct {
	Code     int
	JSONBody interface{}
	FilePath string
}

// JSONTime Json 序列化时用的时间格式
type JSONTime struct {
	time.Time
}

// MarshalJSON 序列化
func (t *JSONTime) MarshalJSON() ([]byte, error) {
	loc, _ := time.LoadLocation("Asia/Shanghai")

	return []byte(fmt.Sprintf(`"%s"`, t.In(loc).Format("2006-01-02 15:04:05"))), nil
}

// UnmarshalJSON 反序列化
func (t *JSONTime) UnmarshalJSON(data []byte) error {
	var err error

	loc, _ := time.LoadLocation("Asia/Shanghai")
	t.Time, err = time.ParseInLocation(`"2006-01-02 15:04:05"`, string(data), loc)
	if err != nil {
		return err
	}

	return nil
}

// Value insert timestamp into mysql need this function.
func (t JSONTime) Value() (driver.Value, error) {
	var zeroTime time.Time
	if t.Time.UnixNano() == zeroTime.UnixNano() {
		return nil, nil
	}
	return t.Time, nil
}

// Scan valueof time.Time
func (t *JSONTime) Scan(v interface{}) error {
	value, ok := v.(time.Time)
	if ok {
		*t = JSONTime{Time: value}
		return nil
	}
	return fmt.Errorf("can not convert %v to timestamp", v)
}

// Fragment 片段，内容片段
type Fragment struct {
	ID        uint      `json:"id" gorm:"primary_key;AUTO_INCREMENT:10000"`
	PrevID    uint      `json:"prevID" gorm:"default:0"`
	Text      string    `json:"text" gorm:"type:text;default:'';not null;"`
	CreatedAt *JSONTime `json:"createdAt,omitempty"`
}

func initFlag() {
	flag.StringVar(&dbPath, "db", "./database", "SQLite3 数据库路径")
	flag.Parse()
	flag.Usage()
}

func initDB() {
	var err error
	sqlite, err = gorm.Open("sqlite3", filepath.Join(dbPath, "solitaire.db"))
	if err != nil {
		panic("failed to connect database: " + err.Error())
	}

	err = sqlite.AutoMigrate(&Fragment{}).Error

	if nil != err {
		panic(err)
	}
}

func main() {
	initFlag()
	initDB()

	defer sqlite.Close()

	router := gin.Default()
	router.Use(func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, ResponseType, accept, origin, Cache-Control, X-Requested-With")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		c.Next()
	})

	v1 := router.Group("/v1")

	v1.GET("/fragments", output(listFragments))
	v1.GET("/fragments/:id", output(listFragments))
	v1.GET("/fragment/:id", output(getFragment))
	v1.POST("/fragment", output(postFragment))

	router.Run(":11823")
}

func output(fn func(*gin.Context) Restful) gin.HandlerFunc {
	return func(c *gin.Context) {
		msg := fn(c)

		if "" != msg.FilePath {
			c.File(msg.FilePath)
		} else {
			c.JSON(msg.Code, msg.JSONBody)
		}
	}
}

func listFragments(c *gin.Context) Restful {
	offset, e := strconv.Atoi(c.DefaultQuery("offset", "0"))

	if nil != e {
		offset = 0
	}

	limit, e := strconv.Atoi(c.DefaultQuery("limit", "10"))

	if nil != e {
		limit = 10
	}

	order, e := strconv.Atoi(c.DefaultQuery("order", "0"))

	if nil != e {
		order = 0
	}

	id, e := strconv.Atoi(c.Param("id"))

	if e != nil {
		id = 0
	}

	var fragments []*Fragment

	tx := sqlite.Where("prev_id==?", id).Limit(limit).Offset(offset)

	if 0 == order {
		e = tx.Order("created_at DESC").Find(&fragments).Error
	} else {
		e = tx.Find(&fragments).Error
	}

	if e != nil {
		return Restful{Code: http.StatusInternalServerError, JSONBody: e.Error}
	} else if 0 == len(fragments) {
		return Restful{Code: http.StatusNotFound, JSONBody: "No fragments"}
	}

	return Restful{Code: http.StatusOK, JSONBody: fragments}
}

func getFragment(c *gin.Context) Restful {
	id, e := strconv.Atoi(c.Param("id"))

	if e != nil {
		return Restful{Code: http.StatusBadRequest, JSONBody: e.Error()}
	}

	var fragment Fragment

	e = sqlite.Where("id==?", id).Find(&fragment).Error

	if e != nil {
		return Restful{Code: http.StatusNotFound, JSONBody: "No fragment"}
	}

	return Restful{Code: http.StatusOK, JSONBody: fragment}
}

func postFragment(c *gin.Context) Restful {
	var fragment Fragment

	e := c.ShouldBindJSON(&fragment)

	if e != nil {
		return Restful{Code: http.StatusBadRequest, JSONBody: "Parsing json to fragment object is abnormal: " + e.Error()}
	}

	e = sqlite.Set("gorm:association_autoupdate", false).Create(&fragment).Error

	if e != nil {
		return Restful{Code: http.StatusInternalServerError, JSONBody: e.Error}
	}

	return Restful{Code: http.StatusOK, JSONBody: fragment.ID}
}
