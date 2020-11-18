package utility

import (
	"bytes"
	"fmt"
	"reflect"
	"strconv"

	log "github.com/cihub/seelog"
)

//TypeToString 几种类型转换为字符串
func TypeToString(v interface{}) string {
	var ret string

	switch paramType := v.(type) {
	case int32:
		ret = strconv.Itoa(int(paramType))
	case uint32:
		ret = strconv.Itoa(int(paramType))
	case int64:
		ret = strconv.Itoa(int(paramType))
	case uint64:
		ret = strconv.Itoa(int(paramType))
	case float32:
		ret = strconv.FormatFloat(float64(paramType), 'f', 6, 32)
	case float64:
		ret = strconv.FormatFloat(float64(paramType), 'f', 6, 64)
	case string:
		ret = paramType
	case bool:
		ret = strconv.FormatBool(paramType)
	case *int32:
		ret = strconv.Itoa(int(*paramType))
	case *uint32:
		ret = strconv.Itoa(int(*paramType))
	case *int64:
		ret = strconv.Itoa(int(*paramType))
	case *uint64:
		ret = strconv.Itoa(int(*paramType))
	case *float32:
		ret = strconv.FormatFloat(float64(*paramType), 'f', 6, 32)
	case *float64:
		ret = strconv.FormatFloat(float64(*paramType), 'f', 6, 64)
	case *string:
		ret = *paramType
	case *bool:
		ret = strconv.FormatBool(*paramType)
	default:
		panic(fmt.Errorf("typeToString unsupport type: %T", v))
	}

	return ret
}

// ConvertTypeToString 几种类型转换为字符串
func ConvertTypeToString(src interface{}) string {
	switch v := src.(type) {
	case string:
		return v
	case []byte:
		return string(v)
	}
	rv := reflect.ValueOf(src)
	switch rv.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return strconv.FormatInt(rv.Int(), 10)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return strconv.FormatUint(rv.Uint(), 10)
	case reflect.Float64:
		return strconv.FormatFloat(rv.Float(), 'f', -1, 64)
	case reflect.Float32:
		return strconv.FormatFloat(rv.Float(), 'f', -1, 32)
	case reflect.Bool:
		return strconv.FormatBool(rv.Bool())
	}
	return fmt.Sprintf("%v", src)
}

// ConvertTypeToString 几种类型转换为字符串
func ConvertReflectVal(src interface{}) string {

	switch v := src.(type) {
	case string:
		return v
	case []byte:
		return string(v)
	}

	return fmt.Sprintf("%v", src)
}

// ArrayToStr 带有0的数组转换为字符串
func ArrayToStr(buff []byte) string {
	index := bytes.IndexByte(buff, 0)
	return string(buff[:index])
}

//Unquote 去掉双引号
func Unquote(str string) string {
	if len(str) > 1 && str[0] == '"' && str[len(str)-1] == '"' {
		return str[1 : len(str)-1]
	}

	return str
}

// Atof 字符串转float
func Atof(value string) float64 {
	if len(value) == 0 {
		return 0.0
	}

	v, err := strconv.ParseFloat(value, 64)
	if err != nil {
		log.Error("value is not float: ", value, ", err: ", err)
		return 0.0
	}

	return v
}

// Atoi 字符串转int
func Atoi(value string) int {
	if len(value) == 0 {
		return 0
	}

	v, err := strconv.Atoi(value)
	if err != nil {
		log.Error("value is not int: ", value, ", err: ", err)
		return 0
	}

	return v
}

// Atob 字符串转bool
func Atob(value string) bool {
	if len(value) == 0 {
		return false
	}

	v, err := strconv.ParseBool(value)
	if err != nil {
		log.Error("value is not bool: ", value, ", err: ", err)
		return false
	}

	return v
}
