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

  public partial class ReadyBattleSpaceResp : TBase
  {
    private string _spaceId;
    private string _uid;
    private bool _ready;

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

    public string Uid
    {
      get
      {
        return _uid;
      }
      set
      {
        __isset.@uid = true;
        this._uid = value;
      }
    }

    public bool Ready
    {
      get
      {
        return _ready;
      }
      set
      {
        __isset.@ready = true;
        this._ready = value;
      }
    }


    public Isset __isset;
    public struct Isset
    {
      public bool spaceId;
      public bool @uid;
      public bool @ready;
    }

    public ReadyBattleSpaceResp()
    {
    }

    public ReadyBattleSpaceResp DeepCopy()
    {
      var tmp77 = new ReadyBattleSpaceResp();
      if((SpaceId != null) && __isset.spaceId)
      {
        tmp77.SpaceId = this.SpaceId;
      }
      tmp77.__isset.spaceId = this.__isset.spaceId;
      if((Uid != null) && __isset.@uid)
      {
        tmp77.Uid = this.Uid;
      }
      tmp77.__isset.@uid = this.__isset.@uid;
      if(__isset.@ready)
      {
        tmp77.Ready = this.Ready;
      }
      tmp77.__isset.@ready = this.__isset.@ready;
      return tmp77;
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
                Uid = await iprot.ReadStringAsync(cancellationToken);
              }
              else
              {
                await TProtocolUtil.SkipAsync(iprot, field.Type, cancellationToken);
              }
              break;
            case 3:
              if (field.Type == TType.Bool)
              {
                Ready = await iprot.ReadBoolAsync(cancellationToken);
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
        var tmp78 = new TStruct("ReadyBattleSpaceResp");
        await oprot.WriteStructBeginAsync(tmp78, cancellationToken);
        var tmp79 = new TField();
        if((SpaceId != null) && __isset.spaceId)
        {
          tmp79.Name = "spaceId";
          tmp79.Type = TType.String;
          tmp79.ID = 1;
          await oprot.WriteFieldBeginAsync(tmp79, cancellationToken);
          await oprot.WriteStringAsync(SpaceId, cancellationToken);
          await oprot.WriteFieldEndAsync(cancellationToken);
        }
        if((Uid != null) && __isset.@uid)
        {
          tmp79.Name = "uid";
          tmp79.Type = TType.String;
          tmp79.ID = 2;
          await oprot.WriteFieldBeginAsync(tmp79, cancellationToken);
          await oprot.WriteStringAsync(Uid, cancellationToken);
          await oprot.WriteFieldEndAsync(cancellationToken);
        }
        if(__isset.@ready)
        {
          tmp79.Name = "ready";
          tmp79.Type = TType.Bool;
          tmp79.ID = 3;
          await oprot.WriteFieldBeginAsync(tmp79, cancellationToken);
          await oprot.WriteBoolAsync(Ready, cancellationToken);
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
      if (!(that is ReadyBattleSpaceResp other)) return false;
      if (ReferenceEquals(this, other)) return true;
      return ((__isset.spaceId == other.__isset.spaceId) && ((!__isset.spaceId) || (global::System.Object.Equals(SpaceId, other.SpaceId))))
        && ((__isset.@uid == other.__isset.@uid) && ((!__isset.@uid) || (global::System.Object.Equals(Uid, other.Uid))))
        && ((__isset.@ready == other.__isset.@ready) && ((!__isset.@ready) || (global::System.Object.Equals(Ready, other.Ready))));
    }

    public override int GetHashCode() {
      int hashcode = 157;
      unchecked {
        if((SpaceId != null) && __isset.spaceId)
        {
          hashcode = (hashcode * 397) + SpaceId.GetHashCode();
        }
        if((Uid != null) && __isset.@uid)
        {
          hashcode = (hashcode * 397) + Uid.GetHashCode();
        }
        if(__isset.@ready)
        {
          hashcode = (hashcode * 397) + Ready.GetHashCode();
        }
      }
      return hashcode;
    }

    public override string ToString()
    {
      var tmp80 = new StringBuilder("ReadyBattleSpaceResp(");
      int tmp81 = 0;
      if((SpaceId != null) && __isset.spaceId)
      {
        if(0 < tmp81++) { tmp80.Append(", "); }
        tmp80.Append("SpaceId: ");
        SpaceId.ToString(tmp80);
      }
      if((Uid != null) && __isset.@uid)
      {
        if(0 < tmp81++) { tmp80.Append(", "); }
        tmp80.Append("Uid: ");
        Uid.ToString(tmp80);
      }
      if(__isset.@ready)
      {
        if(0 < tmp81++) { tmp80.Append(", "); }
        tmp80.Append("Ready: ");
        Ready.ToString(tmp80);
      }
      tmp80.Append(')');
      return tmp80.ToString();
    }
  }

}
