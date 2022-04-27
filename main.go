package main

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

var dbs = make(map[string]string)

type Model struct {
	ID        int    `json:"id"`
	BeginTime string `json:"begin_time"`
	TimeLen   int    `json:"time_len"`
	Condition string `json:"condition"`
	Phone     string `json:"phone"`
	Process   string `json:"process"`
	Nickname  string `json:"nickname"`
	Digest		string `json:"digest"`
	Area       string `json:"area"`
	//id   int    `gorm:"primary_key"`
	//BeginTime string `gorm:"begin_time"`
	//TimeLen  int    `gorm:"time_len"`
	//
}
type Option struct {
	Model
	Description string `json:"description"`
}

type SuggestOption struct {
	ID      int    `json:"id"`
	Suggest string `json:"suggest"`
	Phone   string `json:"phone"`
}

func (Option) TableName() string {
	return "notelist"
}

//type Model struct {
//		id   int    `gorm:"primary_key"`
//		BeginTime string `gorm:"begin_time"`
//		TimeLen  int    `gorm:"time_len"`
//
//}
func getResult(list *[]Option, context *gin.Context, db *gorm.DB) (*gorm.DB, *[]Option) {
	page, _ := strconv.Atoi(context.Query("current"))
	pageSize, _ := strconv.Atoi(context.Query("pageSize"))
	condition := context.Query("condition")
	process := context.Query("process")
	digest := context.Query("digest")
	area :=context.Query("area")
	dbresult := db.Table("notelist").Limit(pageSize).Offset((page - 1) * pageSize).Find(&list)
	fmt.Println(page, pageSize, condition, digest)
	if condition != "" {
		dbresult = dbresult.Where("`condition` = ?", condition).Find(&list)
	}
	if area != "" {
		dbresult = dbresult.Where("`area` = ?", area).Find(&list)
	}
	if process != "" {
		dbresult = dbresult.Where("process = ?", process).Find(&list)

	}
	if digest != "" {
		dbresult = dbresult.Where("description LIKE ?", "%"+digest+"%").Find(&list)
	}
	return dbresult, list
}
func getResultSuggest(list *[]SuggestOption, context *gin.Context, db *gorm.DB) (*gorm.DB, *[]SuggestOption) {
	page, _ := strconv.Atoi(context.Query("current"))
	pageSize, _ := strconv.Atoi(context.Query("pageSize"))
	suggest := context.Query("suggest")
	phone := context.Query("phone")
	dbresult := db.Table("suggesttable").Limit(pageSize).Offset((page - 1) * pageSize).Find(&list)

	if phone != "" {
		dbresult = dbresult.Where("phone = ?", phone).Find(&list)

	}
	if suggest != "" {
		dbresult = dbresult.Where("suggest LIKE ?", "%"+suggest+"%").Find(&list)
	}

	return dbresult, list

}

func setupRouter() *gin.Engine {
	// 1.116.165.233
	dsn := "sjq:root@tcp(43.138.36.76:3306)/notesql?charset=utf8mb4&parseTime=True&loc=Local"
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	//db.AutoMigrate(&Model{})
	if err != nil {
		fmt.Println("sql connect fail")
	}

	// Disable Console Color
	// gin.DisableConsoleColor()
	r := gin.Default()

	// Ping test
	r.GET("/ping", func(c *gin.Context) {
		c.String(http.StatusOK, "pong")
	})

	// Get user value
	r.GET("/user/:name", func(c *gin.Context) {
		user := c.Params.ByName("name")
		value, ok := dbs[user]
		if ok {
			c.JSON(http.StatusOK, gin.H{"user": user, "value": value})
		} else {
			c.JSON(http.StatusOK, gin.H{"user": user, "status": "no value"})
		}
	})

	// Authorized group (uses gin.BasicAuth() middleware)
	// Same than:
	// authorized := r.Group("/")
	// authorized.Use(gin.BasicAuth(gin.Credentials{
	//	  "foo":  "bar",
	//	  "manu": "123",
	//}))
	// authorized := r.Group("/", gin.BasicAuth(gin.Accounts{
	// 	"xiaona": "123",   // user:foo password:bar
	// 	"admin":  "admin", // user:manu password:123
	// }))

	/* example curl for /admin with basicauth header
	   Zm9vOmJhcg== is base64("foo:bar")

		curl -X POST \
	  	http://localhost:8080/admin \
	  	-H 'authorization: Basic Zm9vOmJhcg==' \
	  	-H 'content-type: application/json' \
	  	-d '{"value":"bar"}'
	*/

	r.POST("admin", func(c *gin.Context) {

		user := c.MustGet(gin.AuthUserKey).(string)

		// Parse JSON
		var json struct {
			Value string `json:"value" binding:"required"`
			User  string `json:"user" binding:"required"`
		}
		name := c.PostForm("value")
		//name := c.PostForm("user")
		fmt.Println(name)
		if c.Bind(&json) == nil {
			if json.User == "admin" || json.User == "xiaona" {
				dbs[user] = json.Value
				//c.JSON(http.StatusOK, gin.H{"status": "ok"})
				c.JSON(http.StatusOK, gin.H{"status": "ok", "currentAuthority": json.Value})
			} else {
				c.JSON(http.StatusOK, gin.H{"status": "error", "msg": "用户名密码错误"})
			}

		}

	})

	r.GET("list", func(context *gin.Context) {

		// GoLand是以字母大小写来限定访问域的 只有首字母大写才可以被导出
		//（可以理解为 public ）,子现在我们把 Response 改成这样 ,首字母大写

		type msg struct {
			Status string `json:"status"`
			//List      []Notelist  `gorm:"list"`
			List  *[]Option `json:"list"`
			Total int64     `json:"total"`
		}
		var result msg
		result.Status = "ok"

		var list *[]Option
		var total int64

		db.Table("notelist").Count(&total)
		dbresult, list := getResult(list, context, db)

		if dbresult.Error != nil {
			context.AbortWithStatus(404)
			fmt.Println(err)
		} else {
			result.List = list

			result.Total = total
			context.JSON(200, result)

		}
		//context.JSON(200,result)
	})

	r.POST("getNoteById", func(context *gin.Context) {
		var queryList struct {
			ID int `json:"id" binding:"required"`
		}
		type MSG struct {
			Status string `json:"status"`
			//List      []Notelist  `gorm:"list"`

			Result Option `json:"result"`
		}
		if context.Bind(&queryList) == nil {
			fmt.Println(queryList.ID)
			var msg MSG
			var result Option
			dbresult := db.Table("notelist").First(&result, "id = ?", queryList.ID)
			if dbresult.Error != nil {
				context.AbortWithStatus(404)
				fmt.Println(err)
			} else {
				msg.Status = "ok"
				msg.Result = result
				context.JSON(200, msg)
			}
		}
	})
	r.POST("delNoteById", func(context *gin.Context) {
		var queryList struct {
			ID int `json:"id" binding:"required"`
		}
		type MSG struct {
			Status string `json:"status"`
			//List      []Notelist  `gorm:"list"`
			Msg string `json:"msg"`
		}
		if context.Bind(&queryList) == nil {
			fmt.Println(queryList.ID)
			var msg MSG
			var result Option
			dbresult := db.Delete(&result, queryList.ID)
			if dbresult.Error != nil {
				context.AbortWithStatus(404)
				fmt.Println(err)
			} else {
				msg.Status = "ok"

				msg.Msg = "删除成功"
				context.JSON(200, msg)
			}
		}
	})
	r.POST("editNoteById", func(context *gin.Context) {
		var queryList struct {
			BeginTime   string `json:"begin_time" binding:"required"`
			Description string `json:"description" binding:"required"`
			ID          int    `json:"id`
			TimeLen     int    `json:"time_len" binding:"required"`
			Condition   string `json:"condition"  binding:"required"`
			Phone       string `json:"phone"`
			Process     string `json:"process"`
			Nickname    string `json:"nickname"`
			Digest		string `json:"digest"`
			Area        string `json:"area"`
		}
		type MSG struct {
			Status string `json:"status"`
			//List      []Notelist  `gorm:"list"`
			Msg string `json:"msg"`
		}

		if context.Bind(&queryList) == nil {
			fmt.Println(queryList.ID, queryList.BeginTime, queryList.Description)
			var msg MSG
			var result Option
			var dbresult *gorm.DB
			if queryList.ID != -1 {
				dbresult = db.Where("id = ?", queryList.ID).Take(&result)
				result.BeginTime = queryList.BeginTime
				result.Description = queryList.Description
				result.TimeLen = queryList.TimeLen
				result.ID = queryList.ID
				result.Condition = queryList.Condition
				result.Phone = queryList.Phone
				result.Process = queryList.Process
				result.Nickname = queryList.Nickname
				result.Digest = queryList.Digest
				result.Area = queryList.Area
				db.Save(&result)
				msg.Msg = "编辑成功"
			} else {
				result.BeginTime = queryList.BeginTime
				result.Description = queryList.Description
				result.TimeLen = queryList.TimeLen
				result.Condition = queryList.Condition
				result.Phone = queryList.Phone
				result.Process = queryList.Process
				result.Nickname = queryList.Nickname
				result.Digest = queryList.Digest
				result.Area = queryList.Area
				dbresult = db.Create(&result)
				msg.Msg = "新建成功"
			}
			fmt.Println(queryList.ID, dbresult)

			if dbresult.Error != nil {
				context.AbortWithStatus(404)
				fmt.Println(err)
			} else {
				fmt.Println(result.ID, result.BeginTime, result.Description)

				msg.Status = "ok"
				context.JSON(200, msg)
			}
		}

	})

	r.GET("listsuggest", func(context *gin.Context) {

		// GoLand是以字母大小写来限定访问域的 只有首字母大写才可以被导出
		//（可以理解为 public ）,子现在我们把 Response 改成这样 ,首字母大写

		type msg struct {
			Status string `json:"status"`
			//List      []Notelist  `gorm:"list"`
			List  *[]SuggestOption `json:"list"`
			Total int64            `json:"total"`
		}
		var result msg
		result.Status = "ok"

		var list *[]SuggestOption
		var total int64

		db.Table("`suggesttable`").Count(&total)
		dbresult, list := getResultSuggest(list, context, db)

		if dbresult.Error != nil {
			context.AbortWithStatus(404)
			fmt.Println(err)
		} else {
			result.List = list

			result.Total = total
			context.JSON(200, result)

		}
		//context.JSON(200,result)
	})
	r.POST("editSuggestById", func(context *gin.Context) {
		var queryList struct {
			ID      int    `json:"id`
			Phone   string `json:"phone"`
			Suggest string `json:"suggest"`
		}
		type MSG struct {
			Status string `json:"status"`
			//List      []Notelist  `gorm:"list"`
			Msg string `json:"msg"`
		}

		if context.Bind(&queryList) == nil {

			var msg MSG
			var result SuggestOption
			var dbresult *gorm.DB
			if queryList.ID != -1 {
				dbresult = db.Where("id = ?", queryList.ID).Take(&result)

				result.ID = queryList.ID
				result.Suggest = queryList.Suggest
				result.Phone = queryList.Phone

				db.Table("suggesttable").Save(&result)
				msg.Msg = "编辑成功"
			} else {
				result.Suggest = queryList.Suggest
				result.Phone = queryList.Phone
				dbresult = db.Table("suggesttable").Create(&result)
				msg.Msg = "新建成功"
			}
			fmt.Println(queryList.ID, dbresult)

			if dbresult.Error != nil {
				context.AbortWithStatus(404)
				fmt.Println(err)
			} else {

				msg.Status = "ok"
				context.JSON(200, msg)
			}
		}

	})
	return r
}

func main() {
	r := setupRouter()

	// Listen and Server in 0.0.0.0:8080
	r.Run(":8080")
}
