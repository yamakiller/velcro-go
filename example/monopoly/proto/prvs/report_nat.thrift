namespace go protocols.prvs

struct ReportNat {
    1:string  BattleSpaceID; // 对战空间ID
    2:string  VerifiyCode;   // 对战空间验证码
    3:string  NatAddr;       // Nat地址
}

service ReportNatService{
    ReportNat OnReportNat(1:ReportNat req);
}