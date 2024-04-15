include "login.thrift"
include "../../../../thrift/cluster/prvs/closing.thrift"
namespace go protocols.pubs

service LoginService{
    login.SignInResp  OnSignIn(1:login.SignIn req);
    login.SignOutResp OnSignOut(1:login.SignOut req);
    void OnClientClosed(1:closing.ClientClosed req);
}