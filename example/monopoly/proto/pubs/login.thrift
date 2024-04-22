namespace go  protocols.pubs
namespace netstd  protocols.pubs
struct SignIn{
    1:string token;
}
struct SignInResp{
    1:string Uid;
    2:string DisplayName;
    3:map<string,string> Externs;
}

struct SignOut{
    1:string token;
}

struct SignOutResp{
   1:i32 result;
}