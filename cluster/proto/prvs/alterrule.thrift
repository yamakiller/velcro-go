
include "../../../network/proto/client_id.thrift"

namespace go protocols.prvs
struct RequestGatewayAlterRule {
    1:client_id.ClientID target ;  // 目标(可为空)  
    2:i32            Rule  ;
}

service RequestGatewayAlterRuleService{
    void OnRequestGatewayAlterRule(1: RequestGatewayAlterRule req);
}
