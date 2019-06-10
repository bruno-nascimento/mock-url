package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"strings"

	goEnv "github.com/Netflix/go-env"
	"github.com/go-redis/redis"
	"github.com/kataras/iris"
	"github.com/tidwall/gjson"
	// validator "gopkg.in/go-playground/validator.v9"
)

// https://github.com/kataras/iris/blob/master/_examples/file-server/single-page-application/embedded-single-page-application-with-other-routes/main.go

// Environment from docker-compose
type Environment struct {
	MockURLPort string `env:"MOCK_URL_PORT"`
	RedisURL    string `env:"REDIS_URL"`
	Extras      goEnv.EnvSet
}

// Condition ...
type Condition struct {
	// Matchers []struct {
	// 	FieldQuery         string `json:"fieldQuery"`
	// 	FieldExpectedValue string `json:"fieldExpectedValue"`
	// 	FieldType          string `json:"fieldType"`
	// 	Operator           string `json:"operator"`
	// }
	FieldQuery         string `json:"fieldQuery"`
	FieldExpectedValue string `json:"fieldExpectedValue"`
	FieldType          string `json:"fieldType"`
	Operator           string `json:"operator"`
	Response           string `json:"response"`
}

var redisdb *redis.Client
var env Environment
var operatorFunctionMap = operatorsMap()

func main() {

	es, _ := goEnv.UnmarshalFromEnviron(&env)
	env.Extras = es

	app := iris.Default()
	app.Put("/mock/{method:string}/{path:path}", func(ctx iris.Context) {
		defer ctx.Request().Body.Close()
		rawBodyAsBytes, erro := ioutil.ReadAll(ctx.Request().Body)
		if erro != nil {
			ctx.StatusCode(500)
			ctx.Writef("%v", erro)
			return
		}
		rawBodyAsString := string(rawBodyAsBytes)
		response := gjson.Get(rawBodyAsString, "response").Exists()

		if !response {
			ctx.StatusCode(400)
			ctx.WriteString("You should at least send a body containing a 'response' field. The 'conditions' field is optional.")
			return
		}

		var conditions []Condition
		conditionsJSON := gjson.Get(rawBodyAsString, "conditions")
		_ = json.Unmarshal([]byte(conditionsJSON.Raw), &conditions)
		// fmt.Printf("condition : %#v\n", conditions)
		var conditionErrors = make([]string, 0)
		for index, condition := range conditions {
			var er string
			if len(condition.FieldExpectedValue) < 1 {
				er += " fieldExpectedValue is required;"
			}
			if len(condition.FieldQuery) < 1 {
				er += " fieldQuery is required;"
			}
			if len(condition.FieldType) < 1 {
				er += " fieldType is required;"
			}
			if len(condition.Operator) < 1 {
				er += " operator is required;"
			}
			response := conditionsJSON.Get(fmt.Sprintf("%d", index)).String()
			if len(response) < 1 {
				er += " response is required;"
			} else {
				condition.Response = response
			}
			if len(er) > 0 {
				conditionErrors = append(conditionErrors, er)
			}
		}
		if len(conditionErrors) > 0 {
			ctx.StatusCode(iris.StatusBadRequest)
			ctx.JSON(conditionErrors)
			return
		}

		key := ctx.Params().Get("method") + "-/" + ctx.Params().Get("path")
		set(key, rawBodyAsString)
		ctx.StatusCode(201)
	})
	app.Delete("/mock/{method:string}/{path:path}", func(ctx iris.Context) {
		delete(ctx.Method() + "-" + ctx.Path())
	})
	app.Any("*", func(ctx iris.Context) {
		ctx.Header("X-Mocked-By", "github.com/bruno-nascimento/mock-url")
		value := get(ctx.Method() + "-" + ctx.Path())
		if value == "" {
			ctx.NotFound()
			return
		}
		ctx.ContentType("application/json; charset=UTF-8")
		jsonValue := gjson.Parse(value)
		responseJSON := jsonValue.Get("response").Raw
		if !jsonValue.Get("conditions").Exists() || ctx.Request().ContentLength == 0 {
			ctx.WriteString(responseJSON)
			return
		}
		var conditions []Condition
		conditionsJSON := jsonValue.Get("conditions")
		_ = json.Unmarshal([]byte(conditionsJSON.Raw), &conditions)
		defer ctx.Request().Body.Close()
		rawBodyAsBytes, erro := ioutil.ReadAll(ctx.Request().Body)
		if erro != nil {
			ctx.StatusCode(500)
			ctx.Writef("%#v", erro)
			return
		}
		rawBodyAsString := string(rawBodyAsBytes)
		requestBody := gjson.Parse(rawBodyAsString)
		for index, condition := range conditions {
			queryResult := requestBody.Get(condition.FieldQuery).String()
			if operatorFunctionMap[condition.Operator](queryResult, condition.FieldExpectedValue) {
				ctx.WriteString(conditionsJSON.Get(fmt.Sprintf("%d", index)).Get("response").String())
				return
			}
		}
		ctx.WriteString(responseJSON)
	})
	app.Logger().SetLevel("error")
	app.Run(iris.Addr(env.MockURLPort))
}

func operatorsMap() map[string]func(string, string) bool {
	return map[string]func(string, string) bool{
		"equal": func(p1, p2 string) bool {
			return p1 == p2
		},
		"not_equal": func(p1, p2 string) bool {
			return p1 != p2
		},
	}
}

func client() *redis.Client {
	if redisdb == nil || redisdb.Ping().Err() != nil {
		redisdb = redis.NewClient(&redis.Options{
			Addr: env.RedisURL,
		})
	}
	return redisdb
}

func get(key string) string {
	// fmt.Printf("Chave para recuperar : %s\n", key)
	value, err := client().Do("get", strings.ToLower(key)).String()
	// fmt.Printf("Valor recuperado : %s\n", value)
	if err != nil {
		// fmt.Printf("Erro ao recuperar a chave do redis : %v\n", err)
		return ""
	}
	return value
}

func set(key string, value string) {
	client().Do("set", strings.ToLower(key), value).Result()
	// fmt.Printf("Chave : %s || Valor : %s\n", key, value)
	// result, err := client().Do("set", strings.ToLower(key), value).Result()
	// if err != nil {
	// fmt.Printf("Erro ao inserir chave no redis : %v\n", err)
	// return
	// }
	// fmt.Printf("Chave inserida : %v\n", result)
}

func delete(key string) {
	client().Do("del", strings.ToLower(key))
}
