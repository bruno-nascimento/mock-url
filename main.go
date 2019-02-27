package main

import (
	"io/ioutil"
	"strings"
	"time"

	"github.com/go-redis/redis"
	"github.com/kataras/iris"
)

// https://github.com/kataras/iris/blob/master/_examples/file-server/single-page-application/embedded-single-page-application-with-other-routes/main.go

var redisdb *redis.Client

func main() {
	app := iris.Default()
	app.Put("/mock/{method:string}/{path:path}", func(ctx iris.Context) {
		defer ctx.Request().Body.Close()
		rawBodyAsBytes, err := ioutil.ReadAll(ctx.Request().Body)
		if err != nil {
			ctx.StatusCode(500)
			ctx.Writef("%v", err)
			return
		}
		rawBodyAsString := string(rawBodyAsBytes)
		key := ctx.Params().Get("method") + "-/" + ctx.Params().Get("path")
		setex(key, rawBodyAsString, int(time.Hour*24))
		ctx.WriteString("lero lero " + ctx.Params().Get("method") + "-" + ctx.Params().Get("path"))
	})
	app.Delete("/mock/{method:string}/{path:path}", func(ctx iris.Context) {
		delete(ctx.Method() + "-" + ctx.Path())
	})
	app.Any("*", func(ctx iris.Context) {
		ctx.ContentType("application/json; charset=UTF-8")
		ctx.WriteString(get(ctx.Method() + "-" + ctx.Path()))
	})
	app.Run(iris.Addr(":7777"))
}

func client() *redis.Client {
	if redisdb == nil || redisdb.Ping().Err() != nil {
		redisdb = redis.NewClient(&redis.Options{
			Addr: ":6379",
		})
	}
	return redisdb
}

func get(key string) string {
	value, error := client().Do("get", strings.ToLower(key)).String()
	if error != nil {
		return ""
	}
	return value
}

func setex(key string, value string, seconds int) {
	client().Do("setex", strings.ToLower(key), seconds, value)
}

func delete(key string) {
	client().Do("del", strings.ToLower(key))
}
