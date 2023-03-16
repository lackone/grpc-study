package errcode

var (
	ErrorGetArticleListFail        = NewError(20010001, "获取文章列表失败")
	ErrorGetArticleListRequestFail = NewError(20010002, "获取文章列表请求参数错误")
)
