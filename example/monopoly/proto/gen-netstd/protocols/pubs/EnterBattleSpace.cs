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

  public partial class EnterBattleSpace : TBase
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

    public EnterBattleSpace()
    {
    }

    public EnterBattleSpace DeepCopy()
    {
      var tmp52 = new EnterBattleSpace();
      if((SpaceId != null) && __isset.spaceId)
      {
        tmp52.SpaceId = this.SpaceId;
      }
      tmp52.__isset.spaceId = this.__isset.spaceId;
      return tmp52;
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
        var tmp53 = new TStruct("EnterBattleSpace");
        await oprot.WriteStructBeginAsync(tmp53, cancellationToken);
        var tmp54 = new TField();
        if((SpaceId != null) && __isset.spaceId)
        {
          tmp54.Name = "spaceId";
          tmp54.Type = TType.String;
          tmp54.ID = 1;
          await oprot.WriteFieldBeginAsync(tmp54, cancellationToken);
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
      if (!(that is EnterBattleSpace other)) return false;
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
      var tmp55 = new StringBuilder("EnterBattleSpace(");
      int tmp56 = 0;
      if((SpaceId != null) && __isset.spaceId)
      {
        if(0 < tmp56++) { tmp55.Append(", "); }
        tmp55.Append("SpaceId: ");
        SpaceId.ToString(tmp55);
      }
      tmp55.Append(')');
      return tmp55.ToString();
    }
  }

}
