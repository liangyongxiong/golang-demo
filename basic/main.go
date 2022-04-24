package main

import (
	"context"
	"fmt"
	"os"
	"regexp"
	"strconv"
	"time"

	"github.com/pkg/errors"

	// reflect
	"reflect"

	// flag
	"flag"

	// json
	"encoding/json"

	"github.com/tidwall/gjson"

	// http
	"net/http"
	"net/url"

	"github.com/PuerkitoBio/goquery"
	resty "github.com/go-resty/resty/v2"

	// redis
	"github.com/go-redis/redis"

	// mysql
	"database/sql"

	_ "github.com/go-sql-driver/mysql"
	"github.com/jinzhu/gorm"

	// mongo
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	// log
	rotatelogs "github.com/lestrrat/go-file-rotatelogs"
	"github.com/sirupsen/logrus"

	// excel
	"encoding/csv"

	excelize "github.com/xuri/excelize/v2"

	// progressbar
	progressbar "github.com/schollz/progressbar/v3"
)

func TestRegex() {
	contentLength := int64(753258001)
	header := "Range: bytes=144703488-"
	start, end := int64(0), contentLength
	re := regexp.MustCompile(`bytes=(?P<start>(\d+)?)-(?P<end>(\d+)?)`)
	match := re.FindStringSubmatch(header)
	fmt.Println("----------", match, len(match))
	fmt.Println("----------", re.SubexpNames(), len(re.SubexpNames()), re.SubexpIndex("start"), re.SubexpIndex("end"))
	if len(match) == 0 {
		fmt.Println("invalid byte-range header: %s", header)
		return
	}
	startValue := match[re.SubexpIndex("start")]
	if startValue != "" {
		start, _ = strconv.ParseInt(startValue, 10, 64)
	}
	endValue := match[re.SubexpIndex("end")]
	if endValue != "" {
		end, _ = strconv.ParseInt(endValue, 10, 64)
	}
	fmt.Println("start=", start)
	fmt.Println("end=", end)
}

func TestReflect() {
	now := time.Now()
	typeOfNow := reflect.TypeOf(now).String()
	timestamp := now.Format("2006-01-02 15:04:05") // 固定写法
	typeOfTimestamp := reflect.TypeOf(timestamp)
	fmt.Printf("type of now is %s\n", typeOfNow)
	fmt.Printf("type of timestamp is %s\n", typeOfTimestamp)
}

func TestFlag() {
	namespace := flag.String("namespace", "NAMESPACE", "Specify namespace")
	service := flag.String("service", "SERVICE", "Specify service")
	flag.Parse()
	fmt.Printf("namespace=%s, service=%s\n", *namespace, *service)
}

func TestJson() {
	type Info struct {
		Name string   `json:"name"`
		Say  []string `json:"say"`
	}

	var json_string = []byte(`
		{
			"name" : "Golang",
			"say" : ["Hello", "World!"]
		}
	`)

	var i interface{}
	if err := json.Unmarshal(json_string, &i); err != nil {
		panic(err)
	}
	ninfo := i.(map[string]interface{})
	for k, v := range ninfo {
		fmt.Println(k, v)
	}

	info := Info{}
	if err := json.Unmarshal(json_string, &info); err != nil {
		panic(err)
	}
	fmt.Println(info.Name, info.Say)

	data, err := json.Marshal(info)
	if err != nil {
		panic(err)
	}
	fmt.Println(string(data))

	var json_content = `{
		"name": {"first": "Tom", "last": "Anderson"},
		"friends": [
		  {"first_name": "Dale", "last_name": "Murphy", "age": 44, "nets": ["ig", "fb", "tw"]},
		  {"first_name": "Roger", "last_name": "Craig", "age": 68, "nets": ["fb", "tw"]},
		  {"first_name": "Jane", "last_name": "Murphy", "age": 47, "nets": ["ig", "tw"]}
		]
	}`
	fmt.Printf("all ages of friends: %v\n", gjson.Get(json_content, "friends.#.age"))
	fmt.Printf("age of first friend: %d\n", gjson.Get(json_content, "friends.0.age").Int())

	for _, friend := range gjson.Get(json_content, "friends").Array() {
		fmt.Println(friend)
		fmt.Println(friend.Get("last_name"))
	}
}

func TestHtml() {
	//page := "https://127.0.0.1:3456/"
	page := "https://www.baidu.com/s"
	params := url.Values{
		"ie": {"utf-8"},
		"wd": {"golang"},
	}

	page_with_query := fmt.Sprintf("%s?%s", page, params.Encode())
	request, err := http.NewRequest("GET", page_with_query, nil)
	if err != nil {
		panic(err)
	}

	user_agent := "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/100.0.4896.88 Safari/537.36"
	request.Header.Add("User-Agent", user_agent)

	client := &http.Client{Timeout: 15 * time.Second}
	response, err := client.Do(request)
	if err != nil {
		fmt.Println(errors.WithStack(err))
		return
	}
	defer response.Body.Close()

	doc, _ := goquery.NewDocumentFromReader(response.Body)
	head := doc.Find("head")
	title := head.Find("title")
	fmt.Printf("page title: %s\n", title.Text())

	rest := resty.New()
	resp, _ := rest.R().
		SetQueryParams(map[string]string{
			"ie": "utf-8",
			"wd": "golang",
		}).
		SetHeader("User-Agent", user_agent).
		Get(page)
	fmt.Printf("resty response status: %d\n", resp.StatusCode())
}

func TestProgressbar() {
	bar := progressbar.Default(100)
	for i := 0; i < 100; i++ {
		bar.Add(1)
		time.Sleep(40 * time.Millisecond)
	}
}

func TestRedis() {
	rdb := redis.NewClient(&redis.Options{
		Addr:     "127.0.0.1:6379",
		Password: "", // no password set
		DB:       0,  // use default DB
	})
	defer rdb.Close()

	var err error
	ret, err := rdb.Ping().Result()
	if err != nil {
		fmt.Printf("redis ping failed: %v\n", err)
		return
	}
	fmt.Printf("redis ping success: %v\n", ret)

	err = rdb.Set("score", 100, 0).Err()
	if err != nil {
		fmt.Printf("set score failed, err: %v\n", err)
		return
	}

	scoreStr, err := rdb.Get("score").Result()
	if err != nil {
		panic(err)
	} else if err == redis.Nil {
		fmt.Println("key score does not exist")
		return
	}

	score, err := strconv.Atoi(scoreStr)
	if err != nil {
		panic(err)
	}
	fmt.Printf("get score success: %d\n", score)
}

func TestMysqlRaw() {
	const (
		username = "root"
		password = "root"
		ip       = "127.0.0.1"
		port     = "3306"
		database = "demo"
	)
	source := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8&parseTime=true", username, password, ip, port, database)

	//Db数据库连接池
	db, err := sql.Open("mysql", source)
	if err != nil {
		panic(err)
	}
	defer db.Close()

	//CREATE TABLE `user` (
	//	`id` int NOT NULL AUTO_INCREMENT,
	//	`name` varchar(64) DEFAULT NULL,
	//	`created_at` date DEFAULT NULL,
	//	PRIMARY KEY (`id`)
	//)

	stmt, _ := db.Prepare("INSERT user SET name=?, created_at=?")
	res, _ := stmt.Exec("lyx", time.Now().Format("2006-01-02"))
	id, _ := res.LastInsertId()
	fmt.Printf("Last insert ID: %d\n", id)

	rows, _ := db.Query("SELECT * FROM user")
	for rows.Next() {
		var id int
		var name string
		var created_at *time.Time
		err = rows.Scan(&id, &name, &created_at)
		if err != nil {
			panic(err)
		}
		fmt.Printf("id=%d, name=%s, created_at=%s\n", id, name, created_at.Format("2006-01-02"))
	}
}

func TestMysqlGorm() {
	const (
		username = "root"
		password = "root"
		ip       = "127.0.0.1"
		port     = "3306"
		database = "demo"
	)
	source := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8&parseTime=true", username, password, ip, port, database)

	type User struct {
		Id        int `gorm:"primaryKey"`
		Name      string
		CreatedAt time.Time
	}

	db, err := gorm.Open("mysql", source)
	if err != nil {
		panic(err)
	}
	db.SingularTable(true) // 让grom转义struct名字的时候不用加上s
	defer db.Close()

	var users []User

	db.Find(&users)
	for _, user := range users {
		fmt.Printf("id=%d, name=%s, created_at=%s\n", user.Id, user.Name, user.CreatedAt.Format("2006-01-02"))
	}

	var count int
	db.Model(&User{}).Count(&count)
	fmt.Printf("%d records found\n", count)

	db.Raw(`SELECT * FROM user`).Scan(&users)
	for _, user := range users {
		fmt.Printf("id=%d, name=%s, created_at=%s\n", user.Id, user.Name, user.CreatedAt.Format("2006-01-02"))
	}
}

func TestMongo() {
	const (
		username = "root"
		password = "root"
		ip       = "127.0.0.1"
		port     = "27017"
		database = "demo"
	)

	var err error
	var client *mongo.Client
	uri := fmt.Sprintf("mongodb://%s:%s@%s:%s", username, password, ip, port)
	clientOptions := options.Client().ApplyURI(uri)
	client, err = mongo.Connect(context.TODO(), clientOptions)
	if err != nil {
		panic(err)
	}
	defer client.Disconnect(context.TODO())

	err = client.Ping(context.TODO(), nil)
	if err != nil {
		panic(err)
	}
	fmt.Println("connect to mongodb success")

	// db.users.insert({name: 'lyx', created_at: '2022-04-20'})
	type User struct {
		Name      string
		CreatedAt string
	}

	user := User{Name: "lyx", CreatedAt: "2022-01-01"}
	collection := client.Database(database).Collection("user")
	insertResult, err := collection.InsertOne(context.TODO(), user)
	if err != nil {
		panic(err)
	}
	fmt.Println("Inserted a single document: ", insertResult.InsertedID)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	cur, err := collection.Find(ctx, bson.D{})
	if err != nil {
		panic(err)
	}
	defer cur.Close(ctx)

	for cur.Next(ctx) {
		var result bson.D
		err := cur.Decode(&result)
		if err != nil {
			panic(err)
		}
		fmt.Printf("result.name: %v\n", result.Map()["name"])
	}
	if err := cur.Err(); err != nil {
		panic(err)
	}
}

func TestLog() {
	logrus.WithFields(logrus.Fields{
		"animal": "walrus",
		"size":   10,
	}).Info("A group of walrus emerges from the ocean")

	log := logrus.New()
	fmt.Printf("log level: %d\n", log.Level)
	log.Debug("log debug")
	log.Debugf("log debug f, %d", 10)
	log.Info("log info")
	log.Warn("log warn")
	log.Error("log error")

	log.SetLevel(logrus.DebugLevel)
	log.Formatter = &logrus.TextFormatter{
		DisableColors: true,
		FullTimestamp: true,
	}
	log.Debug("log text formatter test")

	logfile := "./demo.log"
	writer, err := rotatelogs.New(
		logfile+".%Y%m%d",
		rotatelogs.WithLinkName(logfile),          // 生成软链，指向最新日志文件
		rotatelogs.WithMaxAge(7*24*time.Hour),     // 文件最大保存时间
		rotatelogs.WithRotationTime(24*time.Hour), // 日志切割时间间隔
	)
	if err != nil {
		panic(err)
	}

	rotatelog := logrus.New()
	rotatelog.SetOutput(writer)
	rotatelog.Info("hello, world!")
}

func TestExcel() {
	csv_fp, err := os.Create("./demo.csv")
	if err != nil {
		panic(err)
	}
	defer csv_fp.Close()

	csv_fp.WriteString("\xEF\xBB\xBF") // 写入UTF-8 BOM

	writer := csv.NewWriter(csv_fp)
	data := [][]string{
		{"1", "中国", "23"},
		{"2", "美国", "23"},
	}
	writer.WriteAll(data)
	writer.Flush()

	excel_fp := excelize.NewFile()
	index := excel_fp.NewSheet("Sheet2")
	excel_fp.SetCellValue("Sheet2", "A2", "Hello world.")
	excel_fp.SetCellValue("Sheet1", "B2", 100)
	excel_fp.SetActiveSheet(index)
	if err := excel_fp.SaveAs("demo.xlsx"); err != nil {
		fmt.Println(err)
	}
}

func main() {
	TestRegex()
	//TestReflect()
	//TestFlag()
	//TestJson()
	//TestHtml()
	//TestProgressbar()

	//TestRedis()
	//TestMysqlRaw()
	//TestMysqlGorm()
	//TestMongo()
	//TestLog()
	//TestExcel()
}
