package httpclient

import (
    "testing"
    "fmt"
)

type Person struct {
    Mongo []interface{} `json:"mongo"`
    Mysql  string `json:"mysql"`
    Redis string `json:"redis"`
}

func TestPostInf(t *testing.T) {
    var p Person
    //将结果解析成Person结构体
    response, e := PostInf("http://localhost:5000/user-service2/test", nil, nil, &p)
    if e == nil {
        fmt.Println("person: ", response.Data)
    } else {
        panic(e)
    }

    //将结果解析成map
    response, e = PostInf("http://localhost:5000/user-service2/test", nil, nil, nil)
    if e == nil {
        fmt.Println("map: ", response.Data)
    } else {
        panic(e)
    }

    //将结果解析成string
    response, e = PostInf("http://localhost:5000/user-service2/test", nil, nil, "")
    if e == nil {
        fmt.Println("string: ", response.Data)
    } else {
        panic(e)
    }
}

func TestPostJson(t *testing.T) {
    var p Person
    //将结果解析成Person结构体
    response, e := PostJson("http://localhost:5000/user-service2/test", nil, "", &p)
    if e == nil {
        fmt.Println("person: ", response.Data)
    } else {
        panic(e)
    }

    //将结果解析成map
    response, e = PostJson("http://localhost:5000/user-service2/test", nil, "", nil)
    if e == nil {
        fmt.Println("map: ", response.Data)
    } else {
        panic(e)
    }

    //将结果解析成string
    response, e = PostJson("http://localhost:5000/user-service2/test", nil, "", "")
    if e == nil {
        fmt.Println("string: ", response.Data)
    } else {
        panic(e)
    }
}

func TestGetParam(t *testing.T) {
    url := "http://www.baidu.com"
    value := NewFormValue()
    value.Add("matcher[]", "up")
    value.Add("matcher[]", "url_code")
    value.Set("int", "19")
    response, e := GetParam(url, value, "")
    if e != nil {
        panic(e)
    } else {
        fmt.Println(response)
    }
}

func TestPostForm(t *testing.T) {
    maps := make(map[string]string)
    maps["string"] = "hello"
    maps["int"] = "19"
    value := NewFormValue()
    value.Set("string", "hello")
    value.Set("string", "你好")
    value.Add("age", "12")
    response, e := PostForm("https://www.baidu.com", nil, value, "")
    if e!= nil {
        panic(e)
    }
    fmt.Println(response.Data)
}

func TestResultString(t *testing.T) {
    url := "http://www.baidu.com"
    var result string
    Get(url, &result)
    fmt.Println(result)
}

func TestValue(t *testing.T) {
    //headers := Header{}
    //headers.Add("hello", "d")
    //fmt.Println(headers)
}