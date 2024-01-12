using System;
using System.Collections.Generic;
using System.Linq;
using System.Security.Cryptography;
using System.Text;
using System.Threading.Tasks;

namespace Editor.Utils.Encryption
{
    public static class Sha1
    {
        public static string UnicodeComputer(string input)
        {
            SHA1 sha1Hash = SHA1.Create();
            // 将输入字符串转换为字节数组并计算哈希数据  
            byte[] data = sha1Hash.ComputeHash(Encoding.Unicode.GetBytes(input));
            // 创建一个 Stringbuilder 来收集字节并创建字符串  
            StringBuilder str = new StringBuilder();
            // 循环遍历哈希数据的每一个字节并格式化为十六进制字符串  
            for (int i = 0; i < data.Length; i++)
            {
                //加密结果"x2"结果为32位,"x3"结果为48位,"x4"结果为64位
                str.Append(data[i].ToString("x2"));
            }
            // 返回十六进制字符串  
            return str.ToString();
        }

        public static string Utf8Computer(string input)
        {
            SHA1 sha1Hash = SHA1.Create();
            // 将输入字符串转换为字节数组并计算哈希数据  
            byte[] data = sha1Hash.ComputeHash(Encoding.UTF8.GetBytes(input));
            // 创建一个 Stringbuilder 来收集字节并创建字符串  
            StringBuilder str = new StringBuilder();
            // 循环遍历哈希数据的每一个字节并格式化为十六进制字符串  
            for (int i = 0; i < data.Length; i++)
            {
                //加密结果"x2"结果为32位,"x3"结果为48位,"x4"结果为64位
                str.Append(data[i].ToString("x2"));
            }
            // 返回十六进制字符串  
            return str.ToString();
        }
    }
}
