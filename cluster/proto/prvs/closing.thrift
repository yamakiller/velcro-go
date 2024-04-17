include "../../../network/proto/client_id.thrift"
namespace go protocols.prvs
// 请求网关关闭客户端
struct RequestGatewayCloseClient{
    1:client_id.ClientID Target;
}
// 通知某服务客户端连接已关闭
struct ClientClosed{
     1:client_id.ClientID ClientID;
}

service ClientClosedService{
    void OnClientClosed(1:ClientClosed req);
}