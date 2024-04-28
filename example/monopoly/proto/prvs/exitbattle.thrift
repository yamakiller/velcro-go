namespace go protocols.prvs
struct RequestExitBattleSpace {
    1:string  BattleSpaceID;      // 对战空间ID
    2:string  UID ;               // player uid
}

service RequestExitBattleSpaceService{
    RequestExitBattleSpace OnRequestExitBattleSpace(1:RequestExitBattleSpace req);
}