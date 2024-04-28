include "battle.thrift"
namespace go protocols.pubs
namespace netstd  protocols.pubs

service BattleService{
    battle.CreateBattleSpaceResp OnCreateBattleSpace(1:battle.CreateBattleSpace req);
    battle.GetBattleSpaceListResp OnGetBattleSpaceList(1:battle.GetBattleSpaceList req);
    battle.EnterBattleSpaceResp OnEnterBattleSpace(1:battle.EnterBattleSpace req);
    battle.ReadyBattleSpaceResp OnReadyBattleSpace(1:battle.ReadyBattleSpace req);
    battle.RequsetStartBattleSpaceResp OnRequsetStartBattleSpace(1:battle.RequsetStartBattleSpace req);
    battle.ExitBattleSpaceResp OnExitBattleSpace(1:battle.ExitBattleSpace req);
}