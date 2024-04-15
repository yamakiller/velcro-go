include "../../../network/proto/client_id.thrift"
namespace go protocols.prvs
// 服务->网关->客户端
struct ForwardBundle{
  1:  client_id.ClientID sender;// 目标(可为空)  
  2: binary body;// 具体消息
}
// 服务->网关->客户端
struct RequestGatewayPush{
  1:  client_id.ClientID target;// 目标(可为空)  
  2: binary body;// 具体消息
}

service RequestGatewayPushService{
  void OnRequestGatewayPush(1:RequestGatewayPush req);
}