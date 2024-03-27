using Confluent.Kafka;
using System;
using System.Collections.Generic;
using System.ComponentModel;
using System.Configuration;
using System.Diagnostics;
using System.Linq;
using System.Text;
using System.Threading.Tasks;
using System.Windows;
using System.Windows.Threading;
using static Confluent.Kafka.ConfigPropertyNames;

namespace Editor.Utils
{
    public class KafkaConsumer
    {
        public static string brokerUrl = "127.0.0.1:9092";
        public static string topic = "behavior_node";// ConfigurationManager.AppSettings["ConsumeTopic"];
        public static string groupid = "behavior";//  ConfigurationManager.AppSettings["GroupID"];
        public static Action<string>? CallbackEvent = null;
        public static IConsumer<string, string>? consumer = null;
        public static CancellationTokenSource cts;
        private static Thread? _task;
        public static void StartConsume(string url)
        {
            if (url != "")
            {
                brokerUrl = url;
            }
            cts = new CancellationTokenSource();
            _task = new Thread(new ThreadStart(RunConsume));
            _task.Start();
                  
        }
        /// <summary>
        ///     In this example
        ///         - offsets are manually committed.
        ///         - no extra thread is created for the Poll (Consume) loop.
        /// </summary>
        private static void RunConsume()
        {
            var config = new ConsumerConfig
            {
                BootstrapServers = brokerUrl,
                GroupId = groupid,
                AutoOffsetReset = AutoOffsetReset.Latest,
                StatisticsIntervalMs = 0,
                EnableAutoCommit = true
            };

            using (consumer = new ConsumerBuilder<string, string>(config).SetErrorHandler((_, e) => Console.WriteLine($"Error: {e.Reason}")).Build())
            {
                consumer.Subscribe(topic);
                try
                {
                    while (!cts.IsCancellationRequested)
                    {
                        try
                        {
                            var consumeResult = consumer.Consume(cts.Token);
                            Application.Current.Dispatcher.Invoke(() => CallbackEvent?.Invoke(consumeResult.Message.Value));
                           
                            if (!(bool)config.EnableAutoCommit)
                            {
                                consumer.Commit(consumeResult);//手动提交，如果上面的EnableAutoCommit=true表示自动提交，则无需调用Commit方法
                            }
                        }
                        catch(Exception e)
                        {

                        }
                    }
                }
                catch (OperationCanceledException)
                {
                    // Ensure the consumer leaves the group cleanly and final offsets are committed.
                   
                }
            }
        }

        public static void StopConsume()
        {
            try
            {
                cts.Cancel();
                consumer = null;
            }
            catch (Exception)
            {

            }
        }
    }

}
