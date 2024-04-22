
namespace go  protocols.pubs
namespace netstd  protocols.pubs
struct CreateBattleSpace{
    1:string mapURI;
    2:i32 maxCount;
}

struct CreateBattleSpaceResp{
    1:string spaceId;
    2:string mapURI;
}

struct BattleSpacePlayerSimple {
    1:string display;
    2:i32  pos;
}

struct BattleSpaceDataSimple {
    1:string spaceId;
    2:string mapURI;
    3:string masterUid;
    4:string masterIcon;
    5:string masterDisplay;
    6:i32 maxCount;
    7:list<BattleSpacePlayerSimple>players;
}


struct BattleSpacePlayer {
    1:string uid ;
    2:string display;
    3:string icon   ;
    4:i32  pos    ;
}

struct BattleSpaceData {
    1:string  spaceId;
    2:string  mapURI;
    3:string  masterUid ;
    4:i64 starttime;
    5:string   state;
    6:list<BattleSpacePlayer> players;
}
//获取房间列表
struct GetBattleSpaceList {
    1:i32 start;
    2:i32 size;
}

struct GetBattleSpaceListResp {
    1:i32 start;
    2:i32 count;
    3:list< BattleSpaceDataSimple> spaces;
}

//进入房间
struct EnterBattleSpace {
    1:string spaceId;
}

struct EnterBattleSpaceResp {
    1:BattleSpaceData space;
}
//退出房间
struct ExitBattleSpace {
    1:string spaceId;
    2:string uid;
}

struct ExitBattleSpaceResp {
    1:string spaceId;
    2:string uid;
}

//房间准备
struct ReadyBattleSpace {
    1:string spaceId;
    2:string uid;
    3:bool ready;
}

struct ReadyBattleSpaceResp {
    1:string spaceId;
    2:string uid;
    3:bool ready;
}

//解散房间
struct DissBattleSpaceNotify{
    1:string spaceId;
}

//开始战斗
struct RequsetStartBattleSpace{
    1:string spaceId;
}

struct RequsetStartBattleSpaceResp{
    1:string spaceId;
}