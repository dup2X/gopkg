/*
Package errcode 是外部(返回给端的)错误码

  规范如下:

  1.外部错误码常量名为R+大驼峰

  2.外部错误码数值共6位,格式为1-ZZ-Y-XX
       其中ZZ为大类序号,范围[01,99],
       Y为小类序号,范围[1,9],
       XX为具体错误码,范围[01,99]


  3.具体分类如下(!常量名统一采用 以R开头的 大驼峰格式!)
   3.1 成功(0|RSuccess)

   3.2 通用错误码(101yxx|RCommon*)
       3.2.1 参数(1011xx): 参数空、非法、无效等
       3.2.2 服务端(1012xx): 服务器出错、第三方服务异常、数据(库)错误、缓存错误等
       3.2.3 网络/连接(1013xx): 网络不稳定、连接超时等
       3.2.4 签名错误(101401-101420): 验证错误等

   3.14 其他信息(114yxx|RWarn*):匹配不到上面的分类,且不符合独自建立一个分类的


  4.添加错误码时,优先匹配上面的分类(3.1~3.13) 若都不符合,且又不属于单独分类,则请放入3.14.
*/
package errcode

//RespCode : module response code
type RespCode int

/*--------------------------按新规范添加的错误码(如下)--------------------------------*/
const (
	/**********************3.1 成功(0|RSuccess)****************************/

	//RSuccess :成功
	RSuccess RespCode = 0

	/*****************3.2 通用错误码(101yxx|RCommon*)***********************/
	//3.2.1 参数(1011xx): 参数空、非法、无效等

	//RCommonParamEmpty :参数空
	RCommonParamEmpty RespCode = 101101
	//RCommonParamIllegal :参数非法、类型不对
	RCommonParamIllegal RespCode = 101102
	//RCommonParamInvalid :参数无效、不在范围内
	RCommonParamInvalid RespCode = 101103
	//RCommonTokenInvalid :token失效
	RCommonTokenInvalid RespCode = 101104

	//3.2.2 服务端(1012xx): 服务器出错、第三方服务异常、数据(库)错误、缓存错误等

	//RCommonServerError :服务器出错
	RCommonServerError RespCode = 101201
	//RCommonMysqlError :mysql出错
	RCommonMysqlError RespCode = 101202
	//RCommonCurlError :curl出错
	RCommonCurlError RespCode = 101203
	//RCommonThriftError :thrift出错
	RCommonThriftError RespCode = 101204
	//RCommonCacheError :cache出错
	RCommonCacheError RespCode = 101205

	//3.2.3 网络/连接(1013xx): 网络不稳定、连接超时等

	//RCommonNetworkError :网络异常、不稳定
	RCommonNetworkError RespCode = 101301
	//RCommonConnectTimeout :连接超时
	RCommonConnectTimeout RespCode = 101302

	//3.2.4 签名错误(101401-101420): 验证错误等

	//RCommonDsigInvalid :开启验签阻拦后,签名验证错误
	RCommonDsigInvalid RespCode = 101401
)

/*----------------------------现有的旧错误码(如下)-------------------------------------*/
const (
	//Success : 服务器端错误 Server
	Success RespCode = 0

	//SuccessInternalError : internal error
	SuccessInternalError RespCode = -1
	//SuccessJSONDecodeError :json decode error
	SuccessJSONDecodeError RespCode = -2

	// 乘客端 Passenger 错误码

	//PassengerSuccess :成功
	PassengerSuccess RespCode = 0

	//PassengerTokenError :token错误
	PassengerTokenError RespCode = 101
	//PassengerParamsError :参数错误
	PassengerParamsError RespCode = 1010

	//APISigEmpty :api签名为空
	APISigEmpty RespCode = 8001
	//APISigError :api签名有误
	APISigError RespCode = 8002
	//APIReqError :api请求有误
	APIReqError RespCode = 8003
	//APIPhoneTimeError :api请求有误
	APIPhoneTimeError RespCode = 8004
	//APISigTimeout :api签名超时
	APISigTimeout RespCode = 8005
)
