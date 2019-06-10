package main

// import (
// 	"fmt"

// 	"github.com/tidwall/gjson"
// )

const jsonPayload =
// `{
// 	"name": {"first": "Tom", "last": "Anderson"},
// 	"age":37,
// 	"children": ["Sara","Alex","Jack"],
// 	"fav.movie": "Deer Hunter",
// 	"friends":
`[
	  {"first": "Dale", "last": "Murphy", "age": 44},
	  {"first": "Roger", "last": "Craig", "age": 68},
	  {"first": "Jane", "last": "Murphy", "age": 47}
	]`

//   }`

// func contains(slice []string, item string) bool {
// 	set := make(map[string]struct{}, len(slice))
// 	for _, s := range slice {
// 		set[s] = struct{}{}
// 	}

// 	_, ok := set[item]
// 	return ok
// }

// func main() {
// value := gjson.GetMany(jsonPayload, "name.last", "age", "friends.#[last=\"Murphy\"]#.age")
// fmt.Println(value)
// operators := []string{"equal", "not_equal", "bigger", "bigger_or_equal", "smaller", "smaller_or_equal", "contains", "not_contains"}
// fmt.Println(contains(operators, "equal"))
// value := gjson.Get(jsonPayload, "0.first")
// println(value.Str)
// }
