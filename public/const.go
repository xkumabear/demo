package public
//常量定义
const (
	ValidatorKey  = "ValidatorKey"
	TranslatorKey = "TranslatorKey"
	AdminSessionInfoKey = "AdminSessionInfoKey"

	LoadTypeHTTP = 0
	LoadTypeTCP = 1
	LoadTypeGRPC = 2

	HTTPRuleTypePrefixURL = 0
	HTTPRuleTypeDomain = 1
	HTTPNeedHttps = 1
	HTTPNotNeedHttps = 0

	RedisFlowDayKey = "flow_day_count"
	RedisFlowHourKey = "flow_hour_count"

	FlowTotal = "flow_total"   // 全站
	FlowCountServicePrefix= "flow_service"    //服务
	FlowCountAppPrefix = "flow_app"   //租户

	JwtSignKey  =  "my_sign_key"
	JwtExpiresAt = 60*60
)

var  (
	LoadTypeMap = map[int]string{
		LoadTypeHTTP: "HTTP",
		LoadTypeTCP: "TCP",
		LoadTypeGRPC: "GRPC",
	}
)