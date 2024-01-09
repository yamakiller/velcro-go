using System;
using System.Collections.Generic;
using System.Linq;
using System.Text;
using System.Text.Json;
using System.Threading.Tasks;

namespace Editor.Datas
{
    public class Serialization
    {
        public static string Marshal(Workspace workspace)
        {
            var options = new JsonSerializerOptions { WriteIndented = true };
            string jsonString = JsonSerializer.Serialize(workspace, options);

            return jsonString;
        }

        public static Workspace UnMarshal(string  jsonString)
        {
            return JsonSerializer.Deserialize<Workspace>(jsonString);
        }
    }
}
