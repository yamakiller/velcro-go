namespace go ../
struct ClientID{
    1: string Address;  // 地址 
    2: string Id;       // 唯一标记
}
service RpcService{
    void RequestGatewayAlterRule(1: ClientID target,2: i32 Rule);
}
