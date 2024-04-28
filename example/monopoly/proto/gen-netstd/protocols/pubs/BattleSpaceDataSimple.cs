/**
 * <auto-generated>
 * Autogenerated by Thrift Compiler (0.19.0)
 * DO NOT EDIT UNLESS YOU ARE SURE THAT YOU KNOW WHAT YOU ARE DOING
 * </auto-generated>
 */
using System;
using System.Collections;
using System.Collections.Generic;
using System.Text;
using System.IO;
using System.Linq;
using System.Threading;
using System.Threading.Tasks;
using Microsoft.Extensions.Logging;
using Thrift;
using Thrift.Collections;
using Thrift.Protocol;
using Thrift.Protocol.Entities;
using Thrift.Protocol.Utilities;
using Thrift.Transport;
using Thrift.Transport.Client;
using Thrift.Transport.Server;
using Thrift.Processor;


#pragma warning disable IDE0079  // remove unnecessary pragmas
#pragma warning disable IDE0017  // object init can be simplified
#pragma warning disable IDE0028  // collection init can be simplified
#pragma warning disable IDE1006  // parts of the code use IDL spelling
#pragma warning disable CA1822   // empty DeepCopy() methods still non-static
#pragma warning disable IDE0083  // pattern matching "that is not SomeType" requires net5.0 but we still support earlier versions

namespace protocols.pubs
{

  public partial class BattleSpaceDataSimple : TBase
  {
    private string _spaceId;
    private string _mapURI;
    private string _masterUid;
    private string _masterIcon;
    private string _masterDisplay;
    private int _maxCount;
    private List<global::protocols.pubs.BattleSpacePlayerSimple> _players;

    public string SpaceId
    {
      get
      {
        return _spaceId;
      }
      set
      {
        __isset.spaceId = true;
        this._spaceId = value;
      }
    }

    public string MapURI
    {
      get
      {
        return _mapURI;
      }
      set
      {
        __isset.mapURI = true;
        this._mapURI = value;
      }
    }

    public string MasterUid
    {
      get
      {
        return _masterUid;
      }
      set
      {
        __isset.masterUid = true;
        this._masterUid = value;
      }
    }

    public string MasterIcon
    {
      get
      {
        return _masterIcon;
      }
      set
      {
        __isset.masterIcon = true;
        this._masterIcon = value;
      }
    }

    public string MasterDisplay
    {
      get
      {
        return _masterDisplay;
      }
      set
      {
        __isset.masterDisplay = true;
        this._masterDisplay = value;
      }
    }

    public int MaxCount
    {
      get
      {
        return _maxCount;
      }
      set
      {
        __isset.maxCount = true;
        this._maxCount = value;
      }
    }

    public List<global::protocols.pubs.BattleSpacePlayerSimple> Players
    {
      get
      {
        return _players;
      }
      set
      {
        __isset.@players = true;
        this._players = value;
      }
    }


    public Isset __isset;
    public struct Isset
    {
      public bool spaceId;
      public bool mapURI;
      public bool masterUid;
      public bool masterIcon;
      public bool masterDisplay;
      public bool maxCount;
      public bool @players;
    }

    public BattleSpaceDataSimple()
    {
    }

    public BattleSpaceDataSimple DeepCopy()
    {
      var tmp15 = new BattleSpaceDataSimple();
      if((SpaceId != null) && __isset.spaceId)
      {
        tmp15.SpaceId = this.SpaceId;
      }
      tmp15.__isset.spaceId = this.__isset.spaceId;
      if((MapURI != null) && __isset.mapURI)
      {
        tmp15.MapURI = this.MapURI;
      }
      tmp15.__isset.mapURI = this.__isset.mapURI;
      if((MasterUid != null) && __isset.masterUid)
      {
        tmp15.MasterUid = this.MasterUid;
      }
      tmp15.__isset.masterUid = this.__isset.masterUid;
      if((MasterIcon != null) && __isset.masterIcon)
      {
        tmp15.MasterIcon = this.MasterIcon;
      }
      tmp15.__isset.masterIcon = this.__isset.masterIcon;
      if((MasterDisplay != null) && __isset.masterDisplay)
      {
        tmp15.MasterDisplay = this.MasterDisplay;
      }
      tmp15.__isset.masterDisplay = this.__isset.masterDisplay;
      if(__isset.maxCount)
      {
        tmp15.MaxCount = this.MaxCount;
      }
      tmp15.__isset.maxCount = this.__isset.maxCount;
      if((Players != null) && __isset.@players)
      {
        tmp15.Players = this.Players.DeepCopy();
      }
      tmp15.__isset.@players = this.__isset.@players;
      return tmp15;
    }

    public async global::System.Threading.Tasks.Task ReadAsync(TProtocol iprot, CancellationToken cancellationToken)
    {
      iprot.IncrementRecursionDepth();
      try
      {
        TField field;
        await iprot.ReadStructBeginAsync(cancellationToken);
        while (true)
        {
          field = await iprot.ReadFieldBeginAsync(cancellationToken);
          if (field.Type == TType.Stop)
          {
            break;
          }

          switch (field.ID)
          {
            case 1:
              if (field.Type == TType.String)
              {
                SpaceId = await iprot.ReadStringAsync(cancellationToken);
              }
              else
              {
                await TProtocolUtil.SkipAsync(iprot, field.Type, cancellationToken);
              }
              break;
            case 2:
              if (field.Type == TType.String)
              {
                MapURI = await iprot.ReadStringAsync(cancellationToken);
              }
              else
              {
                await TProtocolUtil.SkipAsync(iprot, field.Type, cancellationToken);
              }
              break;
            case 3:
              if (field.Type == TType.String)
              {
                MasterUid = await iprot.ReadStringAsync(cancellationToken);
              }
              else
              {
                await TProtocolUtil.SkipAsync(iprot, field.Type, cancellationToken);
              }
              break;
            case 4:
              if (field.Type == TType.String)
              {
                MasterIcon = await iprot.ReadStringAsync(cancellationToken);
              }
              else
              {
                await TProtocolUtil.SkipAsync(iprot, field.Type, cancellationToken);
              }
              break;
            case 5:
              if (field.Type == TType.String)
              {
                MasterDisplay = await iprot.ReadStringAsync(cancellationToken);
              }
              else
              {
                await TProtocolUtil.SkipAsync(iprot, field.Type, cancellationToken);
              }
              break;
            case 6:
              if (field.Type == TType.I32)
              {
                MaxCount = await iprot.ReadI32Async(cancellationToken);
              }
              else
              {
                await TProtocolUtil.SkipAsync(iprot, field.Type, cancellationToken);
              }
              break;
            case 7:
              if (field.Type == TType.List)
              {
                {
                  var _list16 = await iprot.ReadListBeginAsync(cancellationToken);
                  Players = new List<global::protocols.pubs.BattleSpacePlayerSimple>(_list16.Count);
                  for(int _i17 = 0; _i17 < _list16.Count; ++_i17)
                  {
                    global::protocols.pubs.BattleSpacePlayerSimple _elem18;
                    _elem18 = new global::protocols.pubs.BattleSpacePlayerSimple();
                    await _elem18.ReadAsync(iprot, cancellationToken);
                    Players.Add(_elem18);
                  }
                  await iprot.ReadListEndAsync(cancellationToken);
                }
              }
              else
              {
                await TProtocolUtil.SkipAsync(iprot, field.Type, cancellationToken);
              }
              break;
            default: 
              await TProtocolUtil.SkipAsync(iprot, field.Type, cancellationToken);
              break;
          }

          await iprot.ReadFieldEndAsync(cancellationToken);
        }

        await iprot.ReadStructEndAsync(cancellationToken);
      }
      finally
      {
        iprot.DecrementRecursionDepth();
      }
    }

    public async global::System.Threading.Tasks.Task WriteAsync(TProtocol oprot, CancellationToken cancellationToken)
    {
      oprot.IncrementRecursionDepth();
      try
      {
        var tmp19 = new TStruct("BattleSpaceDataSimple");
        await oprot.WriteStructBeginAsync(tmp19, cancellationToken);
        var tmp20 = new TField();
        if((SpaceId != null) && __isset.spaceId)
        {
          tmp20.Name = "spaceId";
          tmp20.Type = TType.String;
          tmp20.ID = 1;
          await oprot.WriteFieldBeginAsync(tmp20, cancellationToken);
          await oprot.WriteStringAsync(SpaceId, cancellationToken);
          await oprot.WriteFieldEndAsync(cancellationToken);
        }
        if((MapURI != null) && __isset.mapURI)
        {
          tmp20.Name = "mapURI";
          tmp20.Type = TType.String;
          tmp20.ID = 2;
          await oprot.WriteFieldBeginAsync(tmp20, cancellationToken);
          await oprot.WriteStringAsync(MapURI, cancellationToken);
          await oprot.WriteFieldEndAsync(cancellationToken);
        }
        if((MasterUid != null) && __isset.masterUid)
        {
          tmp20.Name = "masterUid";
          tmp20.Type = TType.String;
          tmp20.ID = 3;
          await oprot.WriteFieldBeginAsync(tmp20, cancellationToken);
          await oprot.WriteStringAsync(MasterUid, cancellationToken);
          await oprot.WriteFieldEndAsync(cancellationToken);
        }
        if((MasterIcon != null) && __isset.masterIcon)
        {
          tmp20.Name = "masterIcon";
          tmp20.Type = TType.String;
          tmp20.ID = 4;
          await oprot.WriteFieldBeginAsync(tmp20, cancellationToken);
          await oprot.WriteStringAsync(MasterIcon, cancellationToken);
          await oprot.WriteFieldEndAsync(cancellationToken);
        }
        if((MasterDisplay != null) && __isset.masterDisplay)
        {
          tmp20.Name = "masterDisplay";
          tmp20.Type = TType.String;
          tmp20.ID = 5;
          await oprot.WriteFieldBeginAsync(tmp20, cancellationToken);
          await oprot.WriteStringAsync(MasterDisplay, cancellationToken);
          await oprot.WriteFieldEndAsync(cancellationToken);
        }
        if(__isset.maxCount)
        {
          tmp20.Name = "maxCount";
          tmp20.Type = TType.I32;
          tmp20.ID = 6;
          await oprot.WriteFieldBeginAsync(tmp20, cancellationToken);
          await oprot.WriteI32Async(MaxCount, cancellationToken);
          await oprot.WriteFieldEndAsync(cancellationToken);
        }
        if((Players != null) && __isset.@players)
        {
          tmp20.Name = "players";
          tmp20.Type = TType.List;
          tmp20.ID = 7;
          await oprot.WriteFieldBeginAsync(tmp20, cancellationToken);
          await oprot.WriteListBeginAsync(new TList(TType.Struct, Players.Count), cancellationToken);
          foreach (global::protocols.pubs.BattleSpacePlayerSimple _iter21 in Players)
          {
            await _iter21.WriteAsync(oprot, cancellationToken);
          }
          await oprot.WriteListEndAsync(cancellationToken);
          await oprot.WriteFieldEndAsync(cancellationToken);
        }
        await oprot.WriteFieldStopAsync(cancellationToken);
        await oprot.WriteStructEndAsync(cancellationToken);
      }
      finally
      {
        oprot.DecrementRecursionDepth();
      }
    }

    public override bool Equals(object that)
    {
      if (!(that is BattleSpaceDataSimple other)) return false;
      if (ReferenceEquals(this, other)) return true;
      return ((__isset.spaceId == other.__isset.spaceId) && ((!__isset.spaceId) || (global::System.Object.Equals(SpaceId, other.SpaceId))))
        && ((__isset.mapURI == other.__isset.mapURI) && ((!__isset.mapURI) || (global::System.Object.Equals(MapURI, other.MapURI))))
        && ((__isset.masterUid == other.__isset.masterUid) && ((!__isset.masterUid) || (global::System.Object.Equals(MasterUid, other.MasterUid))))
        && ((__isset.masterIcon == other.__isset.masterIcon) && ((!__isset.masterIcon) || (global::System.Object.Equals(MasterIcon, other.MasterIcon))))
        && ((__isset.masterDisplay == other.__isset.masterDisplay) && ((!__isset.masterDisplay) || (global::System.Object.Equals(MasterDisplay, other.MasterDisplay))))
        && ((__isset.maxCount == other.__isset.maxCount) && ((!__isset.maxCount) || (global::System.Object.Equals(MaxCount, other.MaxCount))))
        && ((__isset.@players == other.__isset.@players) && ((!__isset.@players) || (TCollections.Equals(Players, other.Players))));
    }

    public override int GetHashCode() {
      int hashcode = 157;
      unchecked {
        if((SpaceId != null) && __isset.spaceId)
        {
          hashcode = (hashcode * 397) + SpaceId.GetHashCode();
        }
        if((MapURI != null) && __isset.mapURI)
        {
          hashcode = (hashcode * 397) + MapURI.GetHashCode();
        }
        if((MasterUid != null) && __isset.masterUid)
        {
          hashcode = (hashcode * 397) + MasterUid.GetHashCode();
        }
        if((MasterIcon != null) && __isset.masterIcon)
        {
          hashcode = (hashcode * 397) + MasterIcon.GetHashCode();
        }
        if((MasterDisplay != null) && __isset.masterDisplay)
        {
          hashcode = (hashcode * 397) + MasterDisplay.GetHashCode();
        }
        if(__isset.maxCount)
        {
          hashcode = (hashcode * 397) + MaxCount.GetHashCode();
        }
        if((Players != null) && __isset.@players)
        {
          hashcode = (hashcode * 397) + TCollections.GetHashCode(Players);
        }
      }
      return hashcode;
    }

    public override string ToString()
    {
      var tmp22 = new StringBuilder("BattleSpaceDataSimple(");
      int tmp23 = 0;
      if((SpaceId != null) && __isset.spaceId)
      {
        if(0 < tmp23++) { tmp22.Append(", "); }
        tmp22.Append("SpaceId: ");
        SpaceId.ToString(tmp22);
      }
      if((MapURI != null) && __isset.mapURI)
      {
        if(0 < tmp23++) { tmp22.Append(", "); }
        tmp22.Append("MapURI: ");
        MapURI.ToString(tmp22);
      }
      if((MasterUid != null) && __isset.masterUid)
      {
        if(0 < tmp23++) { tmp22.Append(", "); }
        tmp22.Append("MasterUid: ");
        MasterUid.ToString(tmp22);
      }
      if((MasterIcon != null) && __isset.masterIcon)
      {
        if(0 < tmp23++) { tmp22.Append(", "); }
        tmp22.Append("MasterIcon: ");
        MasterIcon.ToString(tmp22);
      }
      if((MasterDisplay != null) && __isset.masterDisplay)
      {
        if(0 < tmp23++) { tmp22.Append(", "); }
        tmp22.Append("MasterDisplay: ");
        MasterDisplay.ToString(tmp22);
      }
      if(__isset.maxCount)
      {
        if(0 < tmp23++) { tmp22.Append(", "); }
        tmp22.Append("MaxCount: ");
        MaxCount.ToString(tmp22);
      }
      if((Players != null) && __isset.@players)
      {
        if(0 < tmp23++) { tmp22.Append(", "); }
        tmp22.Append("Players: ");
        Players.ToString(tmp22);
      }
      tmp22.Append(')');
      return tmp22.ToString();
    }
  }

}