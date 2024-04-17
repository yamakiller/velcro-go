
include "alterrule.thrift"
include "bundle.thrift"
include "closing.thrift"
namespace go protocols.prvs

service GatewayService{
    bundle.RequestGatewayPush OnRequestGatewayPush(1:bundle.RequestGatewayPush req);
    alterrule.RequestGatewayAlterRule OnRequestGatewayAlterRule(1: alterrule.RequestGatewayAlterRule req);
    closing.RequestGatewayCloseClient OnRequestGatewayCloseClient(1:closing.RequestGatewayCloseClient req);

}