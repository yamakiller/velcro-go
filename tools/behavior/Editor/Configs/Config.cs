using Editor.Datas.Models;
using System;
using System.Collections.Generic;
using System.Linq;
using System.Text;
using System.Threading.Tasks;
using System.Xml.Linq;

namespace Editor.Configs
{
    public static class Config
    {
        public static string DefaultNodeFile { get { return "Resources\\Packages\\default.json"; } }
        public static BehaviorNodeTypeModel UnknowNodeType()
        {
            return new BehaviorNodeTypeModel() { name = "unknow", desc = "新节点", type = "Action" };
        }
    }
}
