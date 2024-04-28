include "login.thrift"
namespace go protocols.pubs
namespace netstd  protocols.pubs
service LoginService{
    login.SignInResp  OnSignIn(1:login.SignIn req);
    login.SignOutResp OnSignOut(1:login.SignOut req);
}