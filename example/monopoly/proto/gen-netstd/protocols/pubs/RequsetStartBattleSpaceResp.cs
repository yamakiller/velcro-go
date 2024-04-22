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

  public partial class RequsetStartBattleSpaceResp : TBase
  {
    private string _spaceId;

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


    public Isset __isset;
    public struct Isset
    {
      public bool spaceId;
    }

    public RequsetStartBattleSpaceResp()
    {
    }

    public RequsetStartBattleSpaceResp DeepCopy()
    {
      var tmp92 = new RequsetStartBattleSpaceResp();
      if((SpaceId != null) && __isset.spaceId)
      {
        tmp92.SpaceId = this.SpaceId;
      }
      tmp92.__isset.spaceId = this.__isset.spaceId;
      return tmp92;
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
        var tmp93 = new TStruct("RequsetStartBattleSpaceResp");
        await oprot.WriteStructBeginAsync(tmp93, cancellationToken);
        var tmp94 = new TField();
        if((SpaceId != null) && __isset.spaceId)
        {
          tmp94.Name = "spaceId";
          tmp94.Type = TType.String;
          tmp94.ID = 1;
          await oprot.WriteFieldBeginAsync(tmp94, cancellationToken);
          await oprot.WriteStringAsync(SpaceId, cancellationToken);
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
      if (!(that is RequsetStartBattleSpaceResp other)) return false;
      if (ReferenceEquals(this, other)) return true;
      return ((__isset.spaceId == other.__isset.spaceId) && ((!__isset.spaceId) || (global::System.Object.Equals(SpaceId, other.SpaceId))));
    }

    public override int GetHashCode() {
      int hashcode = 157;
      unchecked {
        if((SpaceId != null) && __isset.spaceId)
        {
          hashcode = (hashcode * 397) + SpaceId.GetHashCode();
        }
      }
      return hashcode;
    }

    public override string ToString()
    {
      var tmp95 = new StringBuilder("RequsetStartBattleSpaceResp(");
      int tmp96 = 0;
      if((SpaceId != null) && __isset.spaceId)
      {
        if(0 < tmp96++) { tmp95.Append(", "); }
        tmp95.Append("SpaceId: ");
        SpaceId.ToString(tmp95);
      }
      tmp95.Append(')');
      return tmp95.ToString();
    }
  }

}
