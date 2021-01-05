package httpclient

type Response struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data"`
}

/**
*@description:返回错误response
*@param:
*@return:
*@author:吴昊轩
*@time:2019/3/4 0004
 */
func NewErrReponse(err error) Response {
	response := Response{500, err.Error(), nil}
	return response
}

/**
*@description:返回请求成功response
*@param:
*@return:
*@author:吴昊轩
*@time:2019/3/4 0004
 */
func NewSuccessResponse(data interface{}) Response {
	response := Response{0, "success", data}
	return response
}
